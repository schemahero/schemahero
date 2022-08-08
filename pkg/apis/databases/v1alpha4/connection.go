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
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetConnection returns driver name, uri, and any error
func (d Database) GetConnection(ctx context.Context) (string, string, error) {
	isParamBased := false

	// if the connection parameters are not supplied via URI, assume parameter based
	driver, err := d.getDbType()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get database type")
	}

	if driver == "postgres" {
		isParamBased = d.Spec.Connection.Postgres.URI.IsEmpty()
	} else if driver == "cockroachdb" {
		isParamBased = d.Spec.Connection.CockroachDB.URI.IsEmpty()
	} else if driver == "mysql" {
		isParamBased = d.Spec.Connection.Mysql.URI.IsEmpty()
	} else if driver == "cassandra" {
		isParamBased = true
	} else if driver == "rqlite" {
		isParamBased = d.Spec.Connection.RQLite.URI.IsEmpty()
	} else if driver == "timescaledb" {
		isParamBased = d.Spec.Connection.TimescaleDB.URI.IsEmpty()
	}

	if isParamBased {
		return d.getConnectionFromParams(ctx)
	}

	return d.getConnectionFromURI(ctx)
}

func (d Database) getConnectionFromParams(ctx context.Context) (string, string, error) {
	driver, err := d.getDbType()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get database type")
	}

	uri := ""
	if driver == "postgres" {
		hostname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.Host)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres hostname")
		}

		port, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.Port)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres port")
		}

		user, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres user")
		}

		password, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.Password)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres password")
		}

		dbname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.DBName)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres dbname")
		}

		authInfo := url.UserPassword(user, password).String()
		uri = fmt.Sprintf("postgres://%s@%s:%s/%s", authInfo, hostname, port, dbname)

		queryStringCharacter := "?"
		if !d.Spec.Connection.Postgres.SSLMode.IsEmpty() {
			sslMode, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.SSLMode)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read postgres ssl mode")
			}
			uri = fmt.Sprintf("%s%ssslmode=%s", uri, queryStringCharacter, sslMode)
			queryStringCharacter = "&"
		}

		if !d.Spec.Connection.Postgres.CurrentSchema.IsEmpty() {
			currentSchema, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Postgres.CurrentSchema)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read postgres currentSchema")
			}
			uri = fmt.Sprintf("%s%ssearch_path=%s", uri, queryStringCharacter, currentSchema)
			queryStringCharacter = "&"
		}
	} else if driver == "cockroachdb" {
		hostname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.Host)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb hostname")
		}

		port, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.Port)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb port")
		}

		user, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb user")
		}

		password, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.Password)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb password")
		}

		dbname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.DBName)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb dbname")
		}

		authInfo := url.UserPassword(user, password).String()
		uri = fmt.Sprintf("postgres://%s@%s:%s/%s", authInfo, hostname, port, dbname)

		queryStringCharacter := "?"
		if !d.Spec.Connection.CockroachDB.SSLMode.IsEmpty() {
			sslMode, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.SSLMode)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read cockroachdb ssl mode")
			}
			uri = fmt.Sprintf("%s%ssslmode=%s", uri, queryStringCharacter, sslMode)
			queryStringCharacter = "&"
		}

		if !d.Spec.Connection.CockroachDB.CurrentSchema.IsEmpty() {
			currentSchema, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.CockroachDB.CurrentSchema)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read cockroachdb currentSchema")
			}
			uri = fmt.Sprintf("%s%ssearch_path=%s", uri, queryStringCharacter, currentSchema)
			queryStringCharacter = "&"
		}
	} else if driver == "cassandra" {
		return "", "", errors.New("not implemented")
	} else if driver == "mysql" {
		hostname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Mysql.Host)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql hostname")
		}

		port, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Mysql.Port)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql port")
		}

		user, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Mysql.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql user")
		}

		password, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Mysql.Password)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql password")
		}

		dbname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.Mysql.DBName)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql dbname")
		}

		uri = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, hostname, port, dbname)
		if d.Spec.Connection.Mysql.DisableTLS {
			uri = fmt.Sprintf("%s?tls=false", uri)
		}
	} else if driver == "rqlite" {
		hostname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.RQLite.Host)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read rqlite hostname")
		}

		port, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.RQLite.Port)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read rqlite port")
		}

		user, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.RQLite.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read rqlite user")
		}

		password, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.RQLite.Password)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read rqlite password")
		}

		protocol := "https"
		if d.Spec.Connection.RQLite.DisableTLS {
			protocol = "http"
		}
		uri = fmt.Sprintf("%s://%s:%s@%s:%s/", protocol, user, password, hostname, port)
	} else if driver == "timescaledb" {
		hostname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.Host)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read timescale hostname")
		}

		port, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.Port)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read timescale port")
		}

		user, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read timescale user")
		}

		password, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.Password)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read timescale password")
		}

		dbname, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.DBName)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read timescale dbname")
		}

		authInfo := url.UserPassword(user, password).String()
		uri = fmt.Sprintf("postgres://%s@%s:%s/%s", authInfo, hostname, port, dbname)

		queryStringCharacter := "?"
		if !d.Spec.Connection.TimescaleDB.SSLMode.IsEmpty() {
			sslMode, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.SSLMode)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read timescale ssl mode")
			}
			uri = fmt.Sprintf("%s%ssslmode=%s", uri, queryStringCharacter, sslMode)
			queryStringCharacter = "&"
		}

		if !d.Spec.Connection.TimescaleDB.CurrentSchema.IsEmpty() {
			currentSchema, err := d.getValueFromValueOrValueFrom(ctx, driver, d.Spec.Connection.TimescaleDB.CurrentSchema)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read timescale currentSchema")
			}
			uri = fmt.Sprintf("%s%scurrentSchema=%s", uri, queryStringCharacter, currentSchema)
			queryStringCharacter = "&"
		}
	}

	return driver, uri, nil
}

// getConnectionFromURI will return the driver, and a valid connection string for the database. This
// is compatible with any way that the uri was set.
// TODO refactor this to be shorter, simpler and more testable
func (d Database) getConnectionFromURI(ctx context.Context) (string, string, error) {
	driver, err := d.getDbType()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get database type")
	}
	var valueOrValueFrom ValueOrValueFrom
	if driver == "postgres" {
		valueOrValueFrom = d.Spec.Connection.Postgres.URI
	} else if driver == "cockroachdb" {
		valueOrValueFrom = d.Spec.Connection.CockroachDB.URI
	} else if driver == "cassandra" {
		return "", "", errors.New("reading cassandra connecting from uri is not supported")
	} else if driver == "mysql" {
		valueOrValueFrom = d.Spec.Connection.Mysql.URI
	} else if driver == "rqlite" {
		valueOrValueFrom = d.Spec.Connection.RQLite.URI
	} else if driver == "timescaledb" {
		valueOrValueFrom = d.Spec.Connection.TimescaleDB.URI
	}

	value, err := d.getValueFromValueOrValueFrom(ctx, driver, valueOrValueFrom)
	return driver, value, err
}

// getValueFromValueOrValueFrom returns the resolved value, or an error
func (d Database) getValueFromValueOrValueFrom(ctx context.Context, driver string, valueOrValueFrom ValueOrValueFrom) (string, error) {

	// if the value is static, return it
	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	// for other types, we need to talk to the kubernetes api
	cfg, err := config.GetRESTConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed to get config")
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", errors.Wrap(err, "failed to get clientset")
	}

	// if the value is in a secret, look it up and return it
	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secretKeyRefName := valueOrValueFrom.ValueFrom.SecretKeyRef.Name
		secret, err := clientset.CoreV1().Secrets(d.Namespace).Get(ctx, secretKeyRefName, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrap(err, "failed to get secret")
		}
		keyName := valueOrValueFrom.ValueFrom.SecretKeyRef.Key
		keyData, ok := secret.Data[keyName]
		if !ok {
			return "", fmt.Errorf("expected Secret \"%s\" to contain key \"%s\"", secretKeyRefName, keyName)
		}
		return string(keyData), nil
	}

	if valueOrValueFrom.ValueFrom.Vault != nil {
		_, value, err := d.getVaultConnection(ctx, clientset, driver, valueOrValueFrom)
		return value, err
	}

	if valueOrValueFrom.ValueFrom.SSM != nil {
		_, value, err := d.getSSMConnection(ctx, clientset, driver, valueOrValueFrom)
		return value, err
	}

	return "", errors.New("unable to get value for driver")
}
