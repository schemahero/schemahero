package v1alpha4

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetConnectionURIFromTemplate(t *testing.T) {
	tests := []struct {
		name             string
		driver           string
		tmpUser          string
		tmpPassword      string
		valueOrValueFrom ValueOrValueFrom
	}{
		{
			name:        "uses_specified_template",
			driver:      "mysql",
			tmpUser:     "someuser",
			tmpPassword: "p@ssw0rd",
			valueOrValueFrom: ValueOrValueFrom{
				ValueFrom: &ValueFrom{
					Vault: &Vault{
						ConnectionTemplate: "postgresql://{{ .username }}:{{ .password }}@postgresql:5432/schema",
						Secret:             "secret",
					},
				},
			},
		},
	}

	sc := &v1.Secret{}
	sa := &v1.ServiceAccount{
		Secrets: []v1.ObjectReference{
			{
				Namespace: "",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/auth/kubernetes/login":
					_, err := rw.Write([]byte(`{
  "auth": {
    "client_token": "blah"
  }
}`))
					assert.NoError(t, err)
				case "/v1/database/creds/secret":
					_, err := rw.Write([]byte(`{
  "lease_duration": 86400,
  "data": {
    "username": "someuser",
    "password": "p@ssw0rd"
  }
}`))
					assert.NoError(t, err)
				default:
					rw.WriteHeader(http.StatusNotFound)
					_, err := rw.Write([]byte(fmt.Sprintf("Unknown path: %s", r.URL.Path)))
					assert.NoError(t, err)
				}
			}))
			defer srv.Close()

			test.valueOrValueFrom.ValueFrom.Vault.Endpoint = srv.URL

			d := &Database{}
			_, uri, err := d.getVaultConnection(context.TODO(), fake.NewSimpleClientset(sa, sc), test.driver, test.valueOrValueFrom)
			assert.NoError(t, err)

			assert.Equal(t, fmt.Sprintf("postgresql://%s:%s@postgresql:5432/schema", test.tmpUser, test.tmpPassword), uri)
		})
	}
}

func TestGetConnectionURIFromVault(t *testing.T) {
	tests := []struct {
		name             string
		driver           string
		tmpUser          string
		tmpPassword      string
		valueOrValueFrom ValueOrValueFrom
	}{
		{
			name:        "uses_template_from_vault",
			driver:      "mysql",
			tmpUser:     "someuser",
			tmpPassword: "p@ssw0rd",
			valueOrValueFrom: ValueOrValueFrom{
				ValueFrom: &ValueFrom{
					Vault: &Vault{
						Secret: "secret",
					},
				},
			},
		},
	}

	sc := &v1.Secret{}
	sa := &v1.ServiceAccount{
		Secrets: []v1.ObjectReference{
			{
				Namespace: "",
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			d := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db_name",
				},
			}
			s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/auth/kubernetes/login":
					rw.Write([]byte(`{
  "auth": {
    "client_token": "blah"
  }
}`))
				case "/v1/database/creds/secret":
					rw.Write([]byte(`{
  "lease_duration": 86400,
  "data": {
    "username": "someuser",
    "password": "p@ssw0rd"
  }
}`))
				case "/v1/database/config/db_name":
					rw.Write([]byte(`{
  "data": {
    "connection_details": {
	  "connection_url": "postgresql://{{ .username }}:{{ .password }}@postgresql:5432/schema"
	}
  }
}`))
				default:
					rw.WriteHeader(http.StatusNotFound)
					rw.Write([]byte(fmt.Sprintf("Unknown path: %s", r.URL.Path)))
				}
			}))
			defer s.Close()

			test.valueOrValueFrom.ValueFrom.Vault.Endpoint = s.URL

			_, uri, err := d.getVaultConnection(context.TODO(), fake.NewSimpleClientset(sa, sc), test.driver, test.valueOrValueFrom)
			assert.NoError(t, err)

			assert.Equal(t, fmt.Sprintf("postgresql://%s:%s@postgresql:5432/schema", test.tmpUser, test.tmpPassword), uri)
		})
	}
}

func TestKubernetesAuthEndpoint(t *testing.T) {
	tests := []struct {
		name                     string
		kubernetesAuthEndpoint   string
		expectedKubeAuthEndpoint string
		valueOrValueFrom         ValueOrValueFrom
	}{
		{
			name:                     "uses_default_endpoint",
			expectedKubeAuthEndpoint: "/v1/auth/kubernetes/login",
			valueOrValueFrom: ValueOrValueFrom{
				ValueFrom: &ValueFrom{
					Vault: &Vault{
						ConnectionTemplate: "postgresql://{{ .username }}:{{ .password }}@postgresql:5432/schema",
						Secret:             "secret",
					},
				},
			},
		},
		{
			name:                     "uses_specified_endpoint",
			expectedKubeAuthEndpoint: "/v1/auth/kubernetes_custom/login",
			valueOrValueFrom: ValueOrValueFrom{
				ValueFrom: &ValueFrom{
					Vault: &Vault{
						ConnectionTemplate:     "postgresql://{{ .username }}:{{ .password }}@postgresql:5432/schema",
						Secret:                 "secret",
						KubernetesAuthEndpoint: "/v1/auth/kubernetes_custom/login",
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			endpointReached := false
			var actualEndpoint string
			srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				switch path := r.URL.Path; {
				case path == test.expectedKubeAuthEndpoint:
					endpointReached = true
					rw.Write([]byte(`{
  "auth": {
    "client_token": "blah"
  }
}`))
				case path == "/v1/database/creds/secret":
					rw.Write([]byte(`{
  "lease_duration": 86400,
  "data": {
    "username": "someuser",
    "password": "p@ssw0rd"
  }
}`))
				default:
					actualEndpoint = path
				}
			}))

			test.valueOrValueFrom.ValueFrom.Vault.Endpoint = srv.URL

			d := &Database{}
			sc := &v1.Secret{}
			sa := &v1.ServiceAccount{
				Secrets: []v1.ObjectReference{
					{
						Namespace: "",
					},
				},
			}

			_, _, err := d.getVaultConnection(context.TODO(), fake.NewSimpleClientset(sa, sc), "postgresql", test.valueOrValueFrom)

			assert.NoError(t, err)
			assert.True(t, endpointReached, "Expected endpoint %s to be hit, but instead we called %s", test.expectedKubeAuthEndpoint, actualEndpoint)
		})
	}
}
