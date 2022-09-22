package sqlite

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatement(t *testing.T) {
	tests := []struct {
		name               string
		tableSchema        *schemasv1alpha4.SqliteTableSchema
		tableName          string
		expectedStatements []string
	}{
		{
			name: "simple",
			tableSchema: &schemasv1alpha4.SqliteTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.SqliteTableColumn{
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
			tableSchema: &schemasv1alpha4.SqliteTableSchema{
				PrimaryKey: []string{
					"one",
					"two",
				},
				Columns: []*schemasv1alpha4.SqliteTableColumn{
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
						Type: "text",
					},
				},
			},
			tableName: "composite_primary_key",
			expectedStatements: []string{
				`create table "composite_primary_key" ("one" integer, "two" integer, "three" text, primary key ("one", "two"))`,
			},
		},
		{
			name: "composite unique index",
			tableSchema: &schemasv1alpha4.SqliteTableSchema{
				PrimaryKey: []string{
					"one",
				},
				Indexes: []*schemasv1alpha4.SqliteTableIndex{
					{
						Columns:  []string{"two", "three"},
						IsUnique: true,
					},
				},
				Columns: []*schemasv1alpha4.SqliteTableColumn{
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
						Type: "text",
					},
				},
			},
			tableName: "composite_unique_index",
			expectedStatements: []string{
				`create table "composite_unique_index" ("one" integer, "two" integer, "three" text, primary key ("one"))`,
				`create unique index idx_composite_unique_index_two_three on composite_unique_index (two, three)`,
			},
		},
		{
			name: "simple with index",
			tableSchema: &schemasv1alpha4.SqliteTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.SqliteTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
					{
						Name: "email",
						Type: "text",
					},
				},
				Indexes: []*schemasv1alpha4.SqliteTableIndex{
					{
						Name:    "idx_email",
						Columns: []string{"email"},
					},
				},
			},
			tableName: "simple",
			expectedStatements: []string{
				`create table "simple" ("id" integer, "email" text, primary key ("id"))`,
				`create index idx_email on simple (email)`,
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
