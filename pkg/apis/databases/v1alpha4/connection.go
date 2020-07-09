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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetConnection will return a valid connection string for the database. This
// is compatible with any way that the uri was set.
// TODO refactor this to be shorter, simpler and more testable
func (d Database) GetConnection(ctx context.Context) (string, string, error) {
	driver, err := d.getDbType()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get database type")
	}

	var valueOrValueFrom ValueOrValueFrom
	if driver == "postgres" {
		valueOrValueFrom = d.Spec.Connection.Postgres.URI
	} else if driver == "cockroachdb" {
		valueOrValueFrom = d.Spec.Connection.CockroachDB.URI
	} else if driver == "mysql" {
		valueOrValueFrom = d.Spec.Connection.Mysql.URI
	}

	// if the value is static, return it
	if valueOrValueFrom.Value != "" {
		return driver, valueOrValueFrom.Value, nil
	}

	// for other types, we need to talk to the kubernetes api
	cfg, err := config.GetConfig()
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
