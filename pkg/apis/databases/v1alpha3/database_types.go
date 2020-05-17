/*
Copyright 2019 Replicated, Inc.

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

package v1alpha3

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseConnection defines connection parameters for the database driver
type DatabaseConnection struct {
	Postgres    *PostgresConnection    `json:"postgres,omitempty"`
	Mysql       *MysqlConnection       `json:"mysql,omitempty"`
	CockroachDB *CockroachDBConnection `json:"cockroachdb,omitempty"`
}

type SchemaHero struct {
	Image        string            `json:"image,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type DatabaseSpec struct {
	Connection      DatabaseConnection `json:"connection,omitempty"`
	ImmediateDeploy bool               `json:"immediateDeploy,omitempty"`
	SchemaHero      *SchemaHero        `json:"schemahero,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	IsConnected bool   `json:"isConnected"`
	LastPing    string `json:"lastPing"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Database is the Schema for the databases API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec"`
	Status DatabaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

// UsingVault determines whether the specific Database connection is
// configured using a Vault secret
func (d *Database) UsingVault() bool {
	if d.Spec.Connection.CockroachDB != nil {
		return d.Spec.Connection.CockroachDB.URI.HasVaultSecret()
	}
	if d.Spec.Connection.Postgres != nil {
		return d.Spec.Connection.Postgres.URI.HasVaultSecret()
	}
	if d.Spec.Connection.Mysql != nil {
		return d.Spec.Connection.Mysql.URI.HasVaultSecret()
	}
	return false
}

// getVaultDetails finds the specified Vault details if configured for the
// database connection, or returns an error if either the database
// connection or Vault aren't configured
func (d *Database) getVaultDetails() (*Vault, error) {
	if d.Spec.Connection.CockroachDB != nil {
		return d.Spec.Connection.CockroachDB.URI.GetVaultDetails()
	}
	if d.Spec.Connection.Postgres != nil {
		return d.Spec.Connection.Postgres.URI.GetVaultDetails()
	}
	if d.Spec.Connection.Mysql != nil {
		return d.Spec.Connection.Mysql.URI.GetVaultDetails()
	}
	return nil, fmt.Errorf("No database connection configured for database: %s", d.Name)
}

func (d *Database) getDbType() (string, error) {
	if d.Spec.Connection.CockroachDB != nil {
		return "CockroachDB", nil
	}
	if d.Spec.Connection.Postgres != nil {
		return "Postgres", nil
	}
	if d.Spec.Connection.Mysql != nil {
		return "Mysql", nil

	}
	return "", fmt.Errorf("No database connection configured for database: %s", d.Name)
}

// GetVaultAnnotations configures the required Vault annotations to
// work with Vault secret injection, or returns an error if the Database
// is misconfigured for Vault
func (d *Database) GetVaultAnnotations() (map[string]string, error) {
	v, err := d.getVaultDetails()
	t, err := d.getDbType()
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)

	switch t {
	case "Postgres", "CockroachDB":

		t := fmt.Sprintf(`
{{- with secret "%s" -}}
postgres://{{ .Data.username }}:{{ .Data.password }}@postgres:5432/%s{{- end }}`, v.Secret, d.Name)

		annotations["vault.hashicorp.com/agent-inject-template-schemaherouri"] = t
	case "Mysql":

		t := fmt.Sprintf(`
{{- with secret "%s" -}}
{{ .Data.username }}:{{ .Data.password }}@tcp(mysql:3306)/%s{{- end }}`, v.Secret, d.Name)

		annotations["vault.hashicorp.com/agent-inject-template-schemaherouri"] = t
	}

	annotations["vault.hashicorp.com/agent-inject"] = "true"
	annotations["vault.hashicorp.com/agent-inject-secret-schemaherouri"] = v.Secret
	annotations["vault.hashicorp.com/role"] = v.Role

	return annotations, nil
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
