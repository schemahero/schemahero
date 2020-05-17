package table

import (
	"context"
	"errors"
	"fmt"
	"testing"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReadConnectionURI(t *testing.T) {
	tests := []struct {
		name     string
		value    databasesv1alpha3.ValueOrValueFrom
		secret   *corev1.Secret
		expected string
	}{
		{
			name: "Gets URI string",
			value: databasesv1alpha3.ValueOrValueFrom{
				Value: "postgresql://username:password@postgrs:5433/database",
			},
			expected: "postgresql://username:password@postgrs:5433/database",
		},
		{
			name: "Gets URI from k8s secret",
			value: databasesv1alpha3.ValueOrValueFrom{
				ValueFrom: &databasesv1alpha3.ValueFrom{
					SecretKeyRef: &databasesv1alpha3.SecretKeyRef{
						Name: "postgresql-secret",
						Key:  "uri",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{Name: "postgresql-secret"},
				Data: map[string][]byte{
					"uri": []byte("postgresql://username:password@postgrs:5433/database"),
				},
			},
			expected: "postgresql://username:password@postgrs:5433/database",
		},
		{
			name: "Gets URI from Vault volume",
			value: databasesv1alpha3.ValueOrValueFrom{
				ValueFrom: &databasesv1alpha3.ValueFrom{
					Vault: &databasesv1alpha3.Vault{
						Secret: "/vault/creds/schemahero",
						Role:   "databaseName",
					},
				},
			},
			expected: "/vault/creds/schemahero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test
			t.Parallel()

			r := ReconcileTable{
				Client: &stubClient{
					secret: map[types.NamespacedName]*corev1.Secret{
						{Name: "postgresql-secret"}: test.secret,
					},
				},
			}

			uri, err := r.readConnectionURI("", test.value)
			if err != nil {
				t.Fatal(err)
			}

			if uri != test.expected {
				t.Fatalf("Expected: %s, got: %s\n", test.expected, uri)
			}
		})
	}
}

type stubClient struct {
	secret map[types.NamespacedName]*corev1.Secret
}

func (s *stubClient) Status() client.StatusWriter {
	return nil
}

func (s *stubClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	return nil
}

func (s *stubClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	return nil
}

func (s *stubClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return nil
}

func (s *stubClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (s *stubClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (s *stubClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if val, ok := s.secret[key]; ok {
		sec := obj.(*corev1.Secret)
		val.DeepCopyInto(sec)

		return nil
	}

	for k := range s.secret {
		fmt.Println(k)
	}

	return errors.New("Secret not found")
}

func (s *stubClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return nil
}
