package installer

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
)

func migrationsCRDYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(migrationsCRD(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal migrations crd")
	}

	return result.Bytes(), nil
}

func ensureMigrationsCRD(client *kubernetes.Clientset, extensionsClient *extensionsclient.ApiextensionsV1Client) error {
	_, err := extensionsClient.CustomResourceDefinitions().Get("migrations.schemas.schemahero.io", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get migrations crd")
		}

		_, err := extensionsClient.CustomResourceDefinitions().Create(migrationsCRD())
		if err != nil {
			return errors.Wrap(err, "failed to create migrations crd")
		}
	}

	return nil
}

func migrationsCRD() *extensionsv1.CustomResourceDefinition {
	return &extensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "migrations.schemas.schemahero.io",
		},
		Status: extensionsv1.CustomResourceDefinitionStatus{
			StoredVersions: []string{},
			Conditions:     []extensionsv1.CustomResourceDefinitionCondition{},
			AcceptedNames: extensionsv1.CustomResourceDefinitionNames{
				Kind:   "",
				Plural: "",
			},
		},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Group: "schemas.schemahero.io",
			Names: extensionsv1.CustomResourceDefinitionNames{
				Kind:     "Migration",
				ListKind: "MigrationList",
				Plural:   "migrations",
				Singular: "migratino",
			},
			Scope: "Namespaced",
			Versions: []extensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha3",
					Served:  true,
					Storage: true,
					Subresources: &extensionsv1.CustomResourceSubresources{
						Status: &extensionsv1.CustomResourceSubresourceStatus{},
					},
					Schema: &extensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &extensionsv1.JSONSchemaProps{
							Description: "Migration is the Schema for the migrations API",
							Type:        "object",
							Properties: map[string]extensionsv1.JSONSchemaProps{
								"apiVersion": extensionsv1.JSONSchemaProps{
									Type:        "string",
									Description: `APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources`,
								},
								"kind": extensionsv1.JSONSchemaProps{
									Type:        "string",
									Description: `Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds`,
								},
								"metadata": extensionsv1.JSONSchemaProps{
									Type: "object",
								},
								"status": extensionsv1.JSONSchemaProps{
									Type:        "object",
									Required:    []string{},
									Description: "MigrationStatus defines the observed state of Migration",
									Properties:  map[string]extensionsv1.JSONSchemaProps{},
								},
							},
						},
					},
				},
			},
		},
	}
}
