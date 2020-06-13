package v1alpha5

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ValueOrValueFrom struct {
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// HasVaultSecret returns true if the ValueOrValueFrom
// contains a Vault stanza
func (v *ValueOrValueFrom) HasVaultSecret() bool {
	if v.ValueFrom != nil {
		return v.ValueFrom.Vault != nil
	}
	return false
}

// GetVaultDetails returns the configured Vault details for the
// ValueOrValueFrom, or returns error if Vault stanza is missing
func (v *ValueOrValueFrom) GetVaultDetails() (*Vault, error) {
	if v.HasVaultSecret() {
		return v.ValueFrom.Vault, nil
	}

	return nil, errors.New("No Vault secret configured")
}

func (v *ValueOrValueFrom) Read(clientset *kubernetes.Clientset, namespace string) (string, error) {
	if v.Value != "" {
		return v.Value, nil
	}

	if v.ValueFrom == nil {
		return "", errors.New("value and valueFrom cannot both be nil/empty")
	}

	if v.ValueFrom.SecretKeyRef != nil {
		secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), v.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrap(err, "failed to get secret")
		}

		return string(secret.Data[v.ValueFrom.SecretKeyRef.Key]), nil
	}

	if v.ValueFrom.Vault != nil {
		// this feels wrong, but also doesn't make sense to return a
		// a URI ref as a connection URI?
		return "", nil
	}

	return "", errors.New("unable to find supported valueFrom")
}
