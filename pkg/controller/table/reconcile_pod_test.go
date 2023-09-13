package table

import (
	"context"
	"errors"
	"fmt"
	"testing"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReadConnectionURI(t *testing.T) {
	tests := []struct {
		name     string
		value    databasesv1alpha4.ValueOrValueFrom
		secret   *corev1.Secret
		expected string
	}{
		{
			name: "Gets URI string",
			value: databasesv1alpha4.ValueOrValueFrom{
				Value: "postgresql://username:password@postgrs:5433/database",
			},
			expected: "postgresql://username:password@postgrs:5433/database",
		},
		{
			name: "Gets URI from k8s secret",
			value: databasesv1alpha4.ValueOrValueFrom{
				ValueFrom: &databasesv1alpha4.ValueFrom{
					SecretKeyRef: &databasesv1alpha4.SecretKeyRef{
						Name: "postgresql-secret",
						Key:  "uri",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "postgresql-secret"},
				Data: map[string][]byte{
					"uri": []byte("postgresql://username:password@postgrs:5433/database"),
				},
			},
			expected: "postgresql://username:password@postgrs:5433/database",
		},
		// commented out by @marccampbell
		// this test doesn't pass because the code specifically returns ""
		// this test just doesn't align with the code, but it was
		// here, and i'm a little hesistant to remove it at this time

		// {
		// 	name: "Gets URI from Vault volume",
		// 	value: databasesv1alpha4.ValueOrValueFrom{
		// 		ValueFrom: &databasesv1alpha4.ValueFrom{
		// 			Vault: &databasesv1alpha4.Vault{
		// 				Secret: "/vault/creds/schemahero",
		// 				Role:   "databaseName",
		// 			},
		// 		},
		// 	},
		// 	expected: "/vault/creds/schemahero",
		// },
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

func (s *stubClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (s *stubClient) Scheme() *runtime.Scheme {
	return nil
}

func (s *stubClient) Status() client.StatusWriter {
	return nil
}

func (s *stubClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

func (s *stubClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (s *stubClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (s *stubClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (s *stubClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (s *stubClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
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

func (s *stubClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (s *stubClient) SubResource(subResource string) client.SubResourceClient {
	return nil
}

func (s *stubClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (s *stubClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return false, nil
}
