package table

import (
	"context"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileTable) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha4.ValueOrValueFrom) (string, error) {
	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	if valueOrValueFrom.ValueFrom == nil {
		return "", errors.New("value and valueFrom cannot both be nil/empty")
	}

	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Name:      valueOrValueFrom.ValueFrom.SecretKeyRef.Name,
			Namespace: namespace,
		}

		if err := r.Get(context.Background(), secretNamespacedName, secret); err != nil {
			if kuberneteserrors.IsNotFound(err) {
				return "", errors.New("table secret not found")
			}

			return "", errors.Wrap(err, "failed to get existing connection secret")
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	if valueOrValueFrom.ValueFrom.Vault != nil {
		// this feels wrong, but also doesn't make sense to return a
		// a URI ref as a connection URI?
		return "", nil
	}

	return "", errors.New("unable to find supported valueFrom")
}
