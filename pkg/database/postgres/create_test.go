package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatement(t *testing.T) {
	tests := []struct {
		name               string
		tableSchema        *schemasv1alpha4.PostgresqlTableSchema
		tableName          string
		expectedStatements []string
	}{
		{
			name: "simple",
			tableSchema: &schemasv1alpha4.PostgresqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.PostgresqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
			},
			tableName: "simple",
			expectedStatements: []string{
				`create table "simple" ("id" integer, primary key ("id"))`,
			},
		},
		{
			name: "composite primary key",
			tableSchema: &schemasv1alpha4.PostgresqlTableSchema{
				PrimaryKey: []string{
					"one",
					"two",
				},
				Columns: []*schemasv1alpha4.PostgresqlTableColumn{
					{
						Name: "one",
						Type: "integer",
					},
					{
						Name: "two",
						Type: "integer",
					},
					{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName: "composite_primary_key",
			expectedStatements: []string{
				`create table "composite_primary_key" ("one" integer, "two" integer, "three" character varying (255), primary key ("one", "two"))`,
			},
		},
		{
			name: "composite unique index",
			tableSchema: &schemasv1alpha4.PostgresqlTableSchema{
				PrimaryKey: []string{
					"one",
				},
				Indexes: []*schemasv1alpha4.PostgresqlTableIndex{
					{
						Columns:  []string{"two", "three"},
						IsUnique: true,
					},
				},
				Columns: []*schemasv1alpha4.PostgresqlTableColumn{
					{
						Name: "one",
						Type: "integer",
					},
					{
						Name: "two",
						Type: "integer",
					},
					{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName: "composite_unique_index",
			expectedStatements: []string{
				`create table "composite_unique_index" ("one" integer, "two" integer, "three" character varying (255), primary key ("one"), constraint "idx_composite_unique_index_two_three" unique ("two", "three"))`,
			},
		},
		{
			name: "simple with trigger",
			tableSchema: &schemasv1alpha4.PostgresqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.PostgresqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
				Triggers: []*schemasv1alpha4.PostgresqlTableTrigger{
					{
						Name: "tgr",
						Events: []string{
							"after insert",
						},
						ForEachRow:       &trueValue,
						ExecuteProcedure: "test()",
					},
				},
			},
			tableName: "simple",
			expectedStatements: []string{
				`create table "simple" ("id" integer, primary key ("id"))`,
				`create trigger "tgr" after insert on "simple" for each row execute procedure test()`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			createTableStatements, err := CreateTableStatements(test.tableName, test.tableSchema)
			req.NoError(err)

			assert.Equal(t, test.expectedStatements, createTableStatements)
		})
	}
}
