package table

import (
	"testing"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
