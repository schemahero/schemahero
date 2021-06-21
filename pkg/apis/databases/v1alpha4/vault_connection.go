/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha4

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getVaultConnection returns the driver, the resolved URI, or an error
func (d *Database) getVaultConnection(ctx context.Context, clientset kubernetes.Interface, driver string, valueOrValueFrom ValueOrValueFrom) (string, string, error) {
	// if the value is in vault and we are using the vault injector, just read the file
	if valueOrValueFrom.ValueFrom.Vault.AgentInject {
		vaultInjectedFileContents, err := ioutil.ReadFile("/vault/secrets/schemaherouri")
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read vault injected file")
		}

		return driver, string(vaultInjectedFileContents), nil
	}

	// And finally, Vault with native vault integration
	serviceAccountNamespace := valueOrValueFrom.ValueFrom.Vault.ServiceAccountNamespace
	if serviceAccountNamespace == "" {
		serviceAccountNamespace = d.Namespace
	}

	serviceAccount, err := clientset.CoreV1().ServiceAccounts(serviceAccountNamespace).Get(ctx, valueOrValueFrom.ValueFrom.Vault.ServiceAccount, metav1.GetOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get vault service account")
	}

	vaultServiceAccountSecret := serviceAccount.Secrets[0]
	vaultServiceAccountSecretNamespace := vaultServiceAccountSecret.Namespace
	if vaultServiceAccountSecretNamespace == "" {
		vaultServiceAccountSecretNamespace = serviceAccountNamespace
	}
	secret, err := clientset.CoreV1().Secrets(vaultServiceAccountSecretNamespace).Get(ctx, vaultServiceAccountSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get vault service account secret")
	}

	loginPayload := struct {
		Role string `json:"role"`
		JWT  string `json:"jwt"`
	}{
		Role: valueOrValueFrom.ValueFrom.Vault.Role,
		JWT:  string(secret.Data["token"]),
	}
	marshalledLoginBody, err := json.Marshal(loginPayload)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to marshal login payload")
	}

	k8sAuthEndpoint := "/v1/auth/kubernetes/login"
	if valueOrValueFrom.ValueFrom.Vault.KubernetesAuthEndpoint != "" {
		k8sAuthEndpoint = valueOrValueFrom.ValueFrom.Vault.KubernetesAuthEndpoint
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", valueOrValueFrom.ValueFrom.Vault.Endpoint, k8sAuthEndpoint), bytes.NewReader(marshalledLoginBody))
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create login request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to execute login request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.Errorf("unexpected response from vault login: %d", resp.StatusCode)
	}

	type LoginResponseAuth struct {
		ClientToken string `json:"client_token"`
	}
	loginResponse := struct {
		Auth LoginResponseAuth `json:"auth"`
	}{
		Auth: LoginResponseAuth{},
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read response body")
	}
	if err := json.Unmarshal(b, &loginResponse); err != nil {
		return "", "", errors.Wrap(err, "failed to unmarshal login response")
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/v1/database/creds/%s", valueOrValueFrom.ValueFrom.Vault.Endpoint, valueOrValueFrom.ValueFrom.Vault.Secret), nil)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("X-Vault-Token", loginResponse.Auth.ClientToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.Errorf("unexpected response code from database/creds vault request: %d", resp.StatusCode)
	}

	credsResponse := struct {
		LeaseDuration int                    `json:"lease_duration"`
		Data          map[string]interface{} `json:"data"`
	}{
		Data: map[string]interface{}{},
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read body")
	}

	if err := json.Unmarshal(b, &credsResponse); err != nil {
		return "", "", errors.Wrap(err, "failed to unmarshal response")
	}

	uriTemplate, err := getConnectionURITemplate(valueOrValueFrom.ValueFrom.Vault, loginResponse.Auth.ClientToken, d.Name)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get connection URI Template")
	}
	funcMap := template.FuncMap{}
	funcMap["username"] = func() string {
		return credsResponse.Data["username"].(string)
	}
	funcMap["password"] = func() string {
		return credsResponse.Data["password"].(string)
	}

	// with the connection url and the username and password (context), we can build a connection string
	t := template.Must(template.New(fmt.Sprintf("%s/%s/%s", d.Namespace, d.Name, d.ResourceVersion)).Funcs(funcMap).Parse(uriTemplate))
	var connectionURI bytes.Buffer
	if err := t.Execute(&connectionURI, credsResponse.Data); err != nil {
		return "", "", errors.Wrap(err, "failed to render vault connection template")
	}

	return driver, connectionURI.String(), nil
}

// get the uri template from the DB spec if set, otherwise use DB config in Vault
func getConnectionURITemplate(vault *Vault, token string, dbName string) (string, error) {
	if vault.ConnectionTemplate != "" {
		return vault.ConnectionTemplate, nil
	} else {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/database/config/%s", vault.Endpoint, dbName), nil)
		if err != nil {
			return "", errors.Wrap(err, "failed to create request")
		}
		req.Header.Add("X-Vault-Token", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", errors.Wrap(err, "failed to execute request")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", errors.Errorf("unexpected response from vault reading config: %d", resp.StatusCode)
		}

		type ConnectionDetails struct {
			ConnectionURL string `json:"connection_url"`
		}
		type ConfigDataResponse struct {
			ConnectionDetails ConnectionDetails `json:"connection_details"`
		}
		configResponse := struct {
			Data ConfigDataResponse `json:"data"`
		}{
			Data: ConfigDataResponse{},
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "failed to read body")
		}
		if err := json.Unmarshal(b, &configResponse); err != nil {
			return "", errors.Wrap(err, "failed to unmarshal response")
		}

		return configResponse.Data.ConnectionDetails.ConnectionURL, nil
	}
}
