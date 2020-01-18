package postgres

import (
	"testing"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableSchema       *schemasv1alpha3.SQLTableSchema
		tableName         string
		expectedStatement string
	}{
		{
			name: "simple",
			tableSchema: &schemasv1alpha3.SQLTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha3.SQLTableColumn{
					&schemasv1alpha3.SQLTableColumn{
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
			tableSchema: &schemasv1alpha3.SQLTableSchema{
				PrimaryKey: []string{
					"one",
					"two",
				},
				Columns: []*schemasv1alpha3.SQLTableColumn{
					&schemasv1alpha3.SQLTableColumn{
						Name: "one",
						Type: "integer",
					},
					&schemasv1alpha3.SQLTableColumn{
						Name: "two",
						Type: "integer",
					},
					&schemasv1alpha3.SQLTableColumn{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName:         "composite_primary_key",
			expectedStatement: `create table "composite_primary_key" ("one" integer, "two" integer, "three" character varying (255), primary key ("one", "two"))`,
		},
		{
			name: "composite unique index",
			tableSchema: &schemasv1alpha3.SQLTableSchema{
				PrimaryKey: []string{
					"one",
				},
				Indexes: []*schemasv1alpha3.SQLTableIndex{
					&schemasv1alpha3.SQLTableIndex{
						Columns:  []string{"two", "three"},
						IsUnique: true,
					},
				},
				Columns: []*schemasv1alpha3.SQLTableColumn{
					&schemasv1alpha3.SQLTableColumn{
						Name: "one",
						Type: "integer",
					},
					&schemasv1alpha3.SQLTableColumn{
						Name: "two",
						Type: "integer",
					},
					&schemasv1alpha3.SQLTableColumn{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName:         "composite_unique_index",
			expectedStatement: `create table "composite_unique_index" ("one" integer, "two" integer, "three" character varying (255), primary key ("one"), constraint "idx_composite_unique_index_two_three" unique ("two", "three"))`,
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
