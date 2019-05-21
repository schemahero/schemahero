package postgres

import (
	"testing"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableSchema       *schemasv1alpha1.PostgresTableSchema
		tableName         string
		expectedStatement string
	}{
		{
			name: "simple",
			tableSchema: &schemasv1alpha1.PostgresTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha1.PostgresTableColumn{
					&schemasv1alpha1.PostgresTableColumn{
						Name: "id",
						Type: "integer",
					},
				},
			},
			tableName:         "simple",
			expectedStatement: `create table "simple" ("id" integer, primary key ("id"))`,
		},
		{
			name: "composite primary key",
			tableSchema: &schemasv1alpha1.PostgresTableSchema{
				PrimaryKey: []string{
					"one",
					"two",
				},
				Columns: []*schemasv1alpha1.PostgresTableColumn{
					&schemasv1alpha1.PostgresTableColumn{
						Name: "one",
						Type: "integer",
					},
					&schemasv1alpha1.PostgresTableColumn{
						Name: "two",
						Type: "integer",
					},
					&schemasv1alpha1.PostgresTableColumn{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName:         "composite_primary_key",
			expectedStatement: `create table "composite_primary_key" ("one" integer, "two" integer, "three" character varying (255), primary key ("one", "two"))`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			createTableStatement, err := CreateTableStatement(test.tableName, test.tableSchema)
			req.NoError(err)

			assert.Equal(t, test.expectedStatement, createTableStatement)
		})
	}
}
