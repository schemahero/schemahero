package table

import (
	"reflect"
	"testing"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_planConfigMap(t *testing.T) {
	tests := []struct {
		name     string
		table    schemasv1alpha3.Table
		database databasesv1alpha3.Database
		expect   string
	}{
		{
			name: "basic test",
			table: schemasv1alpha3.Table{
				Spec: schemasv1alpha3.TableSpec{
					Database: "db",
					Name:     "name",
					Schema: &schemasv1alpha3.TableSchema{
						Postgres: &schemasv1alpha3.SQLTableSchema{},
					},
				},
			},
			database: databasesv1alpha3.Database{},
			expect: `database: db
name: name
schema:
  postgres: {}
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			actual, err := getPlanConfigMap(&test.database, &test.table)
			req.NoError(err)

			// check some of the fields on the config map
			assert.Equal(t, actual.Data["table.yaml"], test.expect)
		})
	}
}

func Test_vaultAnnotations(t *testing.T) {
	tests := []struct {
		name                string
		expectedAnnotations map[string]string
		expectedArgs        []string
		database            *databasesv1alpha3.Database
		table               *schemasv1alpha3.Table
	}{
		{
			name: "Configures correctly when using Vault",
			expectedAnnotations: map[string]string{
				"vault.hashicorp.com/agent-inject":                      "true",
				"vault.hashicorp.com/agent-inject-secret-schemaherouri": "database/creds/schemahero",
				"vault.hashicorp.com/role":                              "schemahero-plan",
				"vault.hashicorp.com/agent-inject-template-schemaherouri": `
{{- with secret "database/creds/schemahero" -}}
postgres://{{ .Data.username }}:{{ .Data.password }}@postgres:5432/my-database{{- end }}`,
			},
			expectedArgs: []string{
				"plan",
				"--driver",
				"postgres",
				"--spec-file",
				"/specs/table.yaml",
				"--vault-uri-ref",
				"/vault/secrets/schemaherouri",
			},
			database: &databasesv1alpha3.Database{
				TypeMeta:   v1.TypeMeta{APIVersion: "databases.schemahero.io/v1alpha3", Kind: "Database"},
				ObjectMeta: v1.ObjectMeta{Name: "my-database"},
				Spec: databasesv1alpha3.DatabaseSpec{
					Connection: databasesv1alpha3.DatabaseConnection{
						Postgres: &databasesv1alpha3.PostgresConnection{
							URI: databasesv1alpha3.ValueOrValueFrom{
								ValueFrom: &databasesv1alpha3.ValueFrom{
									Vault: &databasesv1alpha3.Vault{
										Secret: "database/creds/schemahero",
										Role:   "schemahero-plan",
									},
								},
							},
						},
					},
				},
				Status: databasesv1alpha3.DatabaseStatus{},
			},
			table: &schemasv1alpha3.Table{
				TypeMeta:   v1.TypeMeta{APIVersion: "schemas.schemahero.io/v1alpha3", Kind: "Table"},
				ObjectMeta: v1.ObjectMeta{Name: "my-table"},
				Spec: schemasv1alpha3.TableSpec{
					Database: "my-database",
					Name:     "my-table",
					Schema: &schemasv1alpha3.TableSchema{
						Postgres: &schemasv1alpha3.SQLTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha3.SQLTableColumn{
								{
									Name: "id",
									Type: "text",
									Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
										NotNull: new(bool),
									},
								},
								{
									Name: "name",
									Type: "text",
									Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
										NotNull: new(bool),
									},
								},
							},
						},
					},
				},
				Status: schemasv1alpha3.TableStatus{},
			},
		},
		{
			name:                "Configures correctly when not using vault",
			expectedAnnotations: nil,
			expectedArgs: []string{
				"plan",
				"--driver",
				"postgres",
				"--spec-file",
				"/specs/table.yaml",
				"--uri",
				"postgres://user:password@postgres:5432/my-database",
			},
			database: &databasesv1alpha3.Database{
				TypeMeta:   v1.TypeMeta{APIVersion: "databases.schemahero.io/v1alpha3", Kind: "Database"},
				ObjectMeta: v1.ObjectMeta{Name: "my-database"},
				Spec: databasesv1alpha3.DatabaseSpec{
					Connection: databasesv1alpha3.DatabaseConnection{
						Postgres: &databasesv1alpha3.PostgresConnection{
							URI: databasesv1alpha3.ValueOrValueFrom{
								Value: "postgres://user:password@postgres:5432/my-database",
							},
						},
					},
				},
				Status: databasesv1alpha3.DatabaseStatus{},
			},
			table: &schemasv1alpha3.Table{
				TypeMeta:   v1.TypeMeta{APIVersion: "schemas.schemahero.io/v1alpha3", Kind: "Table"},
				ObjectMeta: v1.ObjectMeta{Name: "my-table"},
				Spec: schemasv1alpha3.TableSpec{
					Database: "my-database",
					Name:     "my-table",
					Schema: &schemasv1alpha3.TableSchema{
						Postgres: &schemasv1alpha3.SQLTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha3.SQLTableColumn{
								{
									Name: "id",
									Type: "text",
									Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
										NotNull: new(bool),
									},
								},
								{
									Name: "name",
									Type: "text",
									Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
										NotNull: new(bool),
									},
								},
							},
						},
					},
				},
				Status: schemasv1alpha3.TableStatus{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test
			t.Parallel()

			r := &ReconcileTable{}
			actual, err := r.getPlanPod(test.database, test.table)
			if err != nil {
				t.Fatal(err)
			}

			actualAnnotations := actual.ObjectMeta.Annotations
			actualArgs := actual.Spec.Containers[0].Args

			if !reflect.DeepEqual(test.expectedAnnotations, actualAnnotations) {
				t.Fatalf("Expected:\n%s\ngot:\n%s\n", test.expectedAnnotations, actualAnnotations)
			}

			if !reflect.DeepEqual(test.expectedArgs, actualArgs) {
				t.Fatalf("Expected:\n%s\ngot:\n%s\n", test.expectedArgs, actualArgs)
			}
		})
	}
}
