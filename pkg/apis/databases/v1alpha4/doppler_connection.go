package v1alpha4

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const dopplerAPIBaseURL = "https://api.doppler.com"

var dopplerHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type dopplerSecretResponse struct {
	Value struct {
		Computed *string `json:"computed"`
	} `json:"value"`
}

func (d *Database) getDopplerConnection(ctx context.Context, clientset kubernetes.Interface, driver string, valueOrValueFrom *ValueOrValueFrom) (string, string, error) {
	value, err := d.resolveDopplerSecret(ctx, clientset, dopplerHTTPClient, dopplerAPIBaseURL, valueOrValueFrom.ValueFrom.Doppler)
	if err != nil {
		return "", "", err
	}

	return driver, value, nil
}

func (d *Database) resolveDopplerSecret(ctx context.Context, clientset kubernetes.Interface, httpClient *http.Client, baseURL string, doppler *Doppler) (string, error) {
	if doppler == nil {
		return "", errors.New("doppler configuration is required")
	}
	if doppler.Project == "" {
		return "", errors.New("doppler project is required")
	}
	if doppler.Config == "" {
		return "", errors.New("doppler config is required")
	}
	if doppler.Name == "" {
		return "", errors.New("doppler secret name is required")
	}
	if doppler.TokenSecretRef == nil || doppler.TokenSecretRef.Name == "" || doppler.TokenSecretRef.Key == "" {
		return "", errors.New("doppler tokenSecretRef name and key are required")
	}

	secret, err := clientset.CoreV1().Secrets(d.Namespace).Get(ctx, doppler.TokenSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to get doppler token secret")
	}
	tokenBytes, ok := secret.Data[doppler.TokenSecretRef.Key]
	if !ok {
		return "", fmt.Errorf("expected Secret %q to contain key %q", doppler.TokenSecretRef.Name, doppler.TokenSecretRef.Key)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", errors.New("doppler token cannot be empty")
	}

	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse doppler API URL")
	}
	endpoint.Path = "/v3/configs/config/secret"
	query := endpoint.Query()
	query.Set("project", doppler.Project)
	query.Set("config", doppler.Config)
	query.Set("name", doppler.Name)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create doppler request")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to get doppler secret")
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("failed to get doppler secret: API returned %s", resp.Status)
	}

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil {
		return "", errors.Wrap(err, "failed to read doppler response")
	}
	if len(responseBody) > 1<<20 {
		return "", errors.New("doppler response exceeded 1 MiB")
	}

	var result dopplerSecretResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", errors.Wrap(err, "failed to decode doppler response")
	}
	if result.Value.Computed == nil {
		return "", errors.New("doppler response did not contain a computed secret value")
	}

	return *result.Value.Computed, nil
}
