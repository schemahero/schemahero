package v1alpha4

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResolveDopplerSecret(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v3/configs/config/secret", r.URL.Path)
		assert.Equal(t, "payments api", r.URL.Query().Get("project"))
		assert.Equal(t, "production", r.URL.Query().Get("config"))
		assert.Equal(t, "DATABASE_URL", r.URL.Query().Get("name"))
		assert.Equal(t, "Bearer dp.st.test", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		return httpResponse(http.StatusOK, `{"name":"DATABASE_URL","value":{"raw":"${HOST}","computed":"postgres://user:password@db:5432/app"}}`), nil
	})}

	database := &Database{ObjectMeta: metav1.ObjectMeta{Namespace: "production"}}
	clientset := fake.NewClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "doppler-credentials", Namespace: "production"},
		Data:       map[string][]byte{"token": []byte("dp.st.test\n")},
	})

	value, err := database.resolveDopplerSecret(context.Background(), clientset, httpClient, "https://doppler.example", &Doppler{
		Project:        "payments api",
		Config:         "production",
		Name:           "DATABASE_URL",
		TokenSecretRef: &SecretKeyRef{Name: "doppler-credentials", Key: "token"},
	})
	require.NoError(t, err)
	assert.Equal(t, "postgres://user:password@db:5432/app", value)
}

func TestResolveDopplerSecretErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		doppler    *Doppler
		secrets    []corev1.Secret
		statusCode int
		response   string
		expect     string
	}{
		{
			name:   "missing configuration",
			expect: "doppler configuration is required",
		},
		{
			name:    "missing project",
			doppler: validDopplerConfig(),
			expect:  "doppler project is required",
		},
		{
			name: "missing token secret",
			doppler: &Doppler{
				Project:        "project",
				Config:         "config",
				Name:           "DATABASE_URL",
				TokenSecretRef: &SecretKeyRef{Name: "missing", Key: "token"},
			},
			expect: "failed to get doppler token secret",
		},
		{
			name:    "missing token key",
			doppler: validDopplerConfig(),
			secrets: []corev1.Secret{dopplerTokenSecret(map[string][]byte{})},
			expect:  `expected Secret "doppler-credentials" to contain key "token"`,
		},
		{
			name:    "empty token",
			doppler: validDopplerConfig(),
			secrets: []corev1.Secret{dopplerTokenSecret(map[string][]byte{"token": []byte(" \n")})},
			expect:  "doppler token cannot be empty",
		},
		{
			name:       "authentication failure",
			doppler:    validDopplerConfig(),
			secrets:    []corev1.Secret{dopplerTokenSecret(map[string][]byte{"token": []byte("dp.st.secret")})},
			statusCode: http.StatusUnauthorized,
			response:   `{"messages":["token dp.st.secret is invalid"]}`,
			expect:     "API returned 401 Unauthorized",
		},
		{
			name:       "malformed response",
			doppler:    validDopplerConfig(),
			secrets:    []corev1.Secret{dopplerTokenSecret(map[string][]byte{"token": []byte("dp.st.secret")})},
			statusCode: http.StatusOK,
			response:   `{`,
			expect:     "failed to decode doppler response",
		},
		{
			name:       "missing computed value",
			doppler:    validDopplerConfig(),
			secrets:    []corev1.Secret{dopplerTokenSecret(map[string][]byte{"token": []byte("dp.st.secret")})},
			statusCode: http.StatusOK,
			response:   `{"value":{"raw":"secret"}}`,
			expect:     "doppler response did not contain a computed secret value",
		},
		{
			name:       "oversized response",
			doppler:    validDopplerConfig(),
			secrets:    []corev1.Secret{dopplerTokenSecret(map[string][]byte{"token": []byte("dp.st.secret")})},
			statusCode: http.StatusOK,
			response:   strings.Repeat("x", (1<<20)+1),
			expect:     "doppler response exceeded 1 MiB",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.name == "missing project" {
				test.doppler.Project = ""
			}

			httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				statusCode := test.statusCode
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				return httpResponse(statusCode, test.response), nil
			})}

			objects := make([]corev1.Secret, len(test.secrets))
			copy(objects, test.secrets)
			clientset := fake.NewSimpleClientset()
			for i := range objects {
				_, err := clientset.CoreV1().Secrets("production").Create(context.Background(), &objects[i], metav1.CreateOptions{})
				require.NoError(t, err)
			}

			database := &Database{ObjectMeta: metav1.ObjectMeta{Namespace: "production"}}
			_, err := database.resolveDopplerSecret(context.Background(), clientset, httpClient, "https://doppler.example", test.doppler)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expect)
			assert.NotContains(t, err.Error(), "dp.st.secret")
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func httpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func validDopplerConfig() *Doppler {
	return &Doppler{
		Project:        "project",
		Config:         "config",
		Name:           "DATABASE_URL",
		TokenSecretRef: &SecretKeyRef{Name: "doppler-credentials", Key: "token"},
	}
}

func dopplerTokenSecret(data map[string][]byte) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "doppler-credentials", Namespace: "production"},
		Data:       data,
	}
}
