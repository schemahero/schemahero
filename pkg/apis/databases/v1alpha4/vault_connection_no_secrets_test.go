package v1alpha4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetVaultConnectionServiceAccountWithoutSecrets(t *testing.T) {
	// Kubernetes 1.24 and later no longer auto-create a token secret for a
	// ServiceAccount, so ServiceAccount.Secrets can be empty. The native vault
	// integration must return an error instead of indexing an empty slice.
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schemahero",
			Namespace: "default",
		},
	}

	valueOrValueFrom := &ValueOrValueFrom{
		ValueFrom: &ValueFrom{
			Vault: &Vault{
				Secret:         "secret",
				ServiceAccount: "schemahero",
				Endpoint:       "http://vault.example.com",
			},
		},
	}

	d := &Database{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
	}

	_, _, err := d.getVaultConnection(context.TODO(), fake.NewClientset(sa), "postgres", valueOrValueFrom)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no secrets")
}
