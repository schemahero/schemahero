package generate

import (
	"testing"

	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_generateMysql(t *testing.T) {
	tests := []struct {
		name         string
		driver       string
		dbName       string
		table        types.Table
		primaryKey   []string
		foreignKeys  []*types.ForeignKey
		indexes      []*types.Index
		columns      []*types.Column
		expectedYAML string
	}{
		{
			name:   "generating with auto_increment",
			driver: "mysql",
			dbName: "db",
			table: types.Table{
				Name: "simple",
			},
			primaryKey: []string{"id"},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
					Attributes: &types.ColumnAttributes{
						AutoIncrement: &trueValue,
					},
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: db
  name: simple
  schema:
    mysql:
      primaryKey:
      - id
      columns:
      - name: id
        type: integer
        attributes:
          autoIncrement: true
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			actual, err := generateMysqlTableYAML(test.dbName, &test.table, test.primaryKey, test.foreignKeys, test.indexes, test.columns)
			req.NoError(err)
			assert.Equal(t, test.expectedYAML, actual)
		})
	}
}
