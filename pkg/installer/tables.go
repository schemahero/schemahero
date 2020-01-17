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

func tablesCRDYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(tablesCRD(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal tables crd")
	}

	return result.Bytes(), nil
}

func ensureTablesCRD(client *kubernetes.Clientset, extensionsClient *extensionsclient.ApiextensionsV1Client) error {
	_, err := extensionsClient.CustomResourceDefinitions().Get("tables.schemas.schemahero.io", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get tables crd")
		}

		_, err := extensionsClient.CustomResourceDefinitions().Create(tablesCRD())
		if err != nil {
			return errors.Wrap(err, "failed to create tables crd")
		}
	}

	return nil
}

func tablesCRD() *extensionsv1.CustomResourceDefinition {
	return &extensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "tables.schemas.schemahero.io",
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
				Kind:     "Table",
				ListKind: "TableList",
				Plural:   "tables",
				Singular: "table",
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
						OpenAPIV3Schema: tablesOpenAPIV3Schema(),
					},
				},
			},
		},
	}
}

func mysqlSchema() extensionsv1.JSONSchemaProps {
	return extensionsv1.JSONSchemaProps{
		Type:     "object",
		Required: []string{"primaryKey"},
		Properties: map[string]extensionsv1.JSONSchemaProps{
			"columns": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type:     "object",
						Required: []string{"name", "type"},
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"constraints": extensionsv1.JSONSchemaProps{
								Type: "object",
								Properties: map[string]extensionsv1.JSONSchemaProps{
									"notNull": extensionsv1.JSONSchemaProps{
										Type: "boolean",
									},
								},
							},
							"default": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"type": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
						},
					},
				},
			},
			"foreignKeys": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Required: []string{"columns", "references"},
						Type:     "object",
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"columns": extensionsv1.JSONSchemaProps{
								Type: "array",
								Items: &extensionsv1.JSONSchemaPropsOrArray{
									Schema: &extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"onDelete": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"references": extensionsv1.JSONSchemaProps{
								Type:     "object",
								Required: []string{"columns", "table"},
								Properties: map[string]extensionsv1.JSONSchemaProps{
									"columns": extensionsv1.JSONSchemaProps{
										Type: "array",
										Items: &extensionsv1.JSONSchemaPropsOrArray{
											Schema: &extensionsv1.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"table": extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
			"indexes": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type:     "object",
						Required: []string{"columns"},
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"columns": extensionsv1.JSONSchemaProps{
								Type: "array",
								Items: &extensionsv1.JSONSchemaPropsOrArray{
									Schema: &extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
							"isUnique": extensionsv1.JSONSchemaProps{
								Type: "boolean",
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"type": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
						},
					},
				},
			},
			"isDeleted": extensionsv1.JSONSchemaProps{
				Type: "string",
			},
			"primaryKey": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type: "string",
					},
				},
			},
		},
	}
}

func postgresSchema() extensionsv1.JSONSchemaProps {
	return extensionsv1.JSONSchemaProps{
		Type:     "object",
		Required: []string{"primaryKey"},
		Properties: map[string]extensionsv1.JSONSchemaProps{
			"columns": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type:     "object",
						Required: []string{"name", "type"},
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"constraints": extensionsv1.JSONSchemaProps{
								Type: "object",
								Properties: map[string]extensionsv1.JSONSchemaProps{
									"notNull": extensionsv1.JSONSchemaProps{
										Type: "boolean",
									},
								},
							},
							"default": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"type": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
						},
					},
				},
			},
			"foreignKeys": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Required: []string{"columns", "references"},
						Type:     "object",
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"columns": extensionsv1.JSONSchemaProps{
								Type: "array",
								Items: &extensionsv1.JSONSchemaPropsOrArray{
									Schema: &extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"onDelete": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"references": extensionsv1.JSONSchemaProps{
								Type:     "object",
								Required: []string{"columns", "table"},
								Properties: map[string]extensionsv1.JSONSchemaProps{
									"columns": extensionsv1.JSONSchemaProps{
										Type: "array",
										Items: &extensionsv1.JSONSchemaPropsOrArray{
											Schema: &extensionsv1.JSONSchemaProps{
												Type: "string",
											},
										},
									},
									"table": extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
			"indexes": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type:     "object",
						Required: []string{"columns"},
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"columns": extensionsv1.JSONSchemaProps{
								Type: "array",
								Items: &extensionsv1.JSONSchemaPropsOrArray{
									Schema: &extensionsv1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
							"isUnique": extensionsv1.JSONSchemaProps{
								Type: "boolean",
							},
							"name": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
							"type": extensionsv1.JSONSchemaProps{
								Type: "string",
							},
						},
					},
				},
			},
			"isDeleted": extensionsv1.JSONSchemaProps{
				Type: "boolean",
			},
			"primaryKey": extensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &extensionsv1.JSONSchemaPropsOrArray{
					Schema: &extensionsv1.JSONSchemaProps{
						Type: "string",
					},
				},
			},
		},
	}
}

func tablesOpenAPIV3Schema() *extensionsv1.JSONSchemaProps {
	return &extensionsv1.JSONSchemaProps{
		Description: "Table is the Schema for the tables API",
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
				Description: "TableStatus defines the observed state of Table",
				Properties: map[string]extensionsv1.JSONSchemaProps{
					"plans": extensionsv1.JSONSchemaProps{
						Type: "array",
						Items: &extensionsv1.JSONSchemaPropsOrArray{
							Schema: &extensionsv1.JSONSchemaProps{
								Type: "object",
								Properties: map[string]extensionsv1.JSONSchemaProps{
									"approvedAt": extensionsv1.JSONSchemaProps{
										Type:   "integer",
										Format: "int64",
									},
									"ddl": extensionsv1.JSONSchemaProps{
										Type: "string",
									},
									"executedAt": extensionsv1.JSONSchemaProps{
										Type:   "integer",
										Format: "int64",
									},
									"invalidatedAt": extensionsv1.JSONSchemaProps{
										Type:        "integer",
										Format:      "int64",
										Description: "InvalidatedAt is the unix nano timestamp when this plan was determined to be invalid or outdated",
									},
									"name": extensionsv1.JSONSchemaProps{
										Type: "string",
									},
									"plannedAt": extensionsv1.JSONSchemaProps{
										Type:        "integer",
										Format:      "int64",
										Description: "PlannedAt is the unix nano timestamp when the plan was generated",
									},
									"rejectedAt": extensionsv1.JSONSchemaProps{
										Type:   "integer",
										Format: "int64",
									},
								},
							},
						},
					},
				},
			},
			"spec": extensionsv1.JSONSchemaProps{
				Description: "TableSpec defines the desired state of Table",
				Type:        "object",
				Required:    []string{"database", "name", "schema"},
				Properties: map[string]extensionsv1.JSONSchemaProps{
					"database": extensionsv1.JSONSchemaProps{
						Type: "string",
					},
					"name": extensionsv1.JSONSchemaProps{
						Type: "string",
					},
					"requires": extensionsv1.JSONSchemaProps{
						Type: "array",
						Items: &extensionsv1.JSONSchemaPropsOrArray{
							Schema: &extensionsv1.JSONSchemaProps{
								Type: "string",
							},
						},
					},
					"schema": extensionsv1.JSONSchemaProps{
						Type: "object",
						Properties: map[string]extensionsv1.JSONSchemaProps{
							"mysql":    mysqlSchema(),
							"postgres": postgresSchema(),
						},
					},
				},
			},
		},
	}
}
