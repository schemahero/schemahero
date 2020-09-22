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
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get config")
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get clientset")
	}

	uri := ""
	if driver == "postgres" {
		hostname, err := d.Spec.Connection.Postgres.Host.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres hostname")
		}

		port, err := d.Spec.Connection.Postgres.Port.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres port")
		}

		user, err := d.Spec.Connection.Postgres.User.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres user")
		}

		password, err := d.Spec.Connection.Postgres.Password.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres password")
		}

		dbname, err := d.Spec.Connection.Postgres.DBName.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read postgres dbname")
		}

		uri = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, hostname, port, dbname)
		if !d.Spec.Connection.Postgres.SSLMode.IsEmpty() {
			sslMode, err := d.Spec.Connection.Postgres.SSLMode.Read(clientset, d.Namespace)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read postgres ssl mode")
			}
			uri = fmt.Sprintf("%s?sslmode=%s", uri, sslMode)
		}
	} else if driver == "cockroachdb" {
		hostname, err := d.Spec.Connection.CockroachDB.Host.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb hostname")
		}

		port, err := d.Spec.Connection.CockroachDB.Port.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb port")
		}

		user, err := d.Spec.Connection.CockroachDB.User.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb user")
		}

		password, err := d.Spec.Connection.CockroachDB.Password.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb password")
		}

		dbname, err := d.Spec.Connection.CockroachDB.DBName.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read cockroachdb dbname")
		}

		uri = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, hostname, port, dbname)
		if !d.Spec.Connection.CockroachDB.SSLMode.IsEmpty() {
			sslMode, err := d.Spec.Connection.CockroachDB.SSLMode.Read(clientset, d.Namespace)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read cockroachdb ssl mode")
			}
			uri = fmt.Sprintf("%s?sslmode=%s", uri, sslMode)
		}
	} else if driver == "cassandra" {
		return "", "", errors.New("not implemented")
	} else if driver == "mysql" {
		hostname, err := d.Spec.Connection.Mysql.Host.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql hostname")
		}

		port, err := d.Spec.Connection.Mysql.Port.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql port")
		}

		user, err := d.Spec.Connection.Mysql.User.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql user")
		}

		password, err := d.Spec.Connection.Mysql.Password.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql password")
		}

		dbname, err := d.Spec.Connection.Mysql.DBName.Read(clientset, d.Namespace)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read mysql dbname")
		}

		uri = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, hostname, port, dbname)
		if d.Spec.Connection.Mysql.DisableTLS {
			uri = fmt.Sprintf("%s?tls=false", uri)
		}
	}

	return driver, uri, nil
}

// getConnectionFromURI will return a valid connection string for the database. This
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
	}

	// if the value is static, return it
	if valueOrValueFrom.Value != "" {
		return driver, valueOrValueFrom.Value, nil
	}

	// for other types, we need to talk to the kubernetes api
	cfg, err := config.GetRESTConfig()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get config")
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get clientset")
	}

	// if the value is in a secret, look it up and return it
	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret, err := clientset.CoreV1().Secrets(d.Namespace).Get(ctx, valueOrValueFrom.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return "", "", errors.Wrap(err, "failed to get secret")
		}

		return driver, string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	if valueOrValueFrom.ValueFrom.Vault != nil {
		return d.getVaultConnection(ctx, clientset, driver, valueOrValueFrom)
	}

	if valueOrValueFrom.ValueFrom.SSM != nil {
		return d.getSSMConnection(ctx, clientset, driver, valueOrValueFrom)
	}

	return "", "", errors.New("unable to get connection")
}
