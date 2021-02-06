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

package v1alpha4

import (
	"fmt"

	"github.com/pkg/errors"
)

// UsingVault determines whether the specific Database connection is
// configured using a Vault secret
func (d Database) UsingVault() bool {
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
func (d Database) getVaultDetails() (*Vault, error) {
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

func (d Database) getDbType() (string, error) {
	if d.Spec.Connection.CockroachDB != nil {
		return "cockroachdb", nil
	}
	if d.Spec.Connection.Cassandra != nil {
		return "cassandra", nil
	}
	if d.Spec.Connection.Postgres != nil {
		return "postgres", nil
	}
	if d.Spec.Connection.Mysql != nil {
		return "mysql", nil
	}
	return "", fmt.Errorf("No database connection configured for database: %s", d.Name)
}

// GetVaultAnnotations configures the required Vault annotations to
// work with Vault secret injection, or returns an error if the Database
// is misconfigured for Vault.
// If there is a connection template specified it will be used otherwise
// a default is provided for the supported database types.
func (d *Database) GetVaultAnnotations() (map[string]string, error) {
	if !d.UsingVault() {
		return nil, nil
	}

	v, err := d.getVaultDetails()
	if err != nil {
		return nil, err
	}

	if !v.AgentInject {
		return nil, nil
	}

	db, err := d.getDbType()
	if err != nil {
		return nil, err
	}

	connTemplate := v.ConnectionTemplate

	var template string
	if connTemplate != "" {
		template = fmt.Sprintf(`
{{- with secret "database/creds/%s" -}}
%s{{- end }}`, v.Role, connTemplate)

	} else {
		switch db {
		case "postgres", "cockroachdb":
			template = fmt.Sprintf(`
{{- with secret "database/creds/%s" -}}
postgres://{{ .Data.username }}:{{ .Data.password }}@postgres:5432/%s{{- end }}`, v.Role, d.Name)

		case "mysql":
			template = fmt.Sprintf(`
{{- with secret "database/creds/%s" -}}
{{ .Data.username }}:{{ .Data.password }}@tcp(mysql:3306)/%s{{- end }}`, v.Role, d.Name)

		case "cassandra":
			return nil, errors.New("not implemented")
		default:
			return nil, fmt.Errorf("unknown database driver: %q", db)
		}
	}

	a := map[string]string{
		"vault.hashicorp.com/agent-inject-template-schemaherouri": template,
		"vault.hashicorp.com/agent-inject":                        "true",
		"vault.hashicorp.com/agent-inject-secret-schemaherouri":   fmt.Sprintf("database/creds/%s", v.Role),
		"vault.hashicorp.com/role":                                v.Role,
	}

	return a, nil
}
