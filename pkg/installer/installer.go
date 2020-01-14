package installer

import (
	"github.com/pkg/errors"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func InstallOperator() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes config")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create new kubernetes client")
	}

	extensionsClient, err := extensionsclient.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "faield to create extensions client")
	}

	if err := createDatabasesCRD(client, extensionsClient); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	return nil
}

func createDatabasesCRD(client *kubernetes.Clientset, extensionsClient *extensionsclient.ApiextensionsV1Client) error {
	databasesCRD := extensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "databases.databases.schemahero.io",
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
			Group: "databases.schemahero.io",
			Names: extensionsv1.CustomResourceDefinitionNames{
				Kind:     "Database",
				ListKind: "DatabaseList",
				Plural:   "databases",
				Singular: "database",
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
							Description: "Database is the Schema for the databases API",
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
									Required:    []string{"isConnected", "lastPing"},
									Description: "DatabaseStatus defines the observed state of Database",
									Properties: map[string]extensionsv1.JSONSchemaProps{
										"isConnected": extensionsv1.JSONSchemaProps{
											Type: "boolean",
										},
										"lastPing": extensionsv1.JSONSchemaProps{
											Type: "string",
										},
									},
								},
								"connection":      extensionsv1.JSONSchemaProps{},
								"immediateDeploy": extensionsv1.JSONSchemaProps{},
								"schemahero":      extensionsv1.JSONSchemaProps{},
							},
						},
					},
				},
			},
		},
	}

	if _, err := extensionsClient.CustomResourceDefinitions().Create(&databasesCRD); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	return nil
}
