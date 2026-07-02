package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
)

func Test_AddIndexStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		schema            string
		schemaIndex       *schemasv1alpha4.PostgresqlTableIndex
		expectedStatement string
	}{
		{
			name:      "no name, one column, not specified unique",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
			},
			expectedStatement: `create index idx_t2_c1 on t2 (c1)`,
		},
		{
			name:      "specified name, one column, not specified unique",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
				Name: "idx_name",
			},
			expectedStatement: `create index idx_name on t2 (c1)`,
		},
		{
			name:      "no name, two columns, not specified unique",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
					"c2",
				},
			},
			expectedStatement: `create index idx_t2_c1_c2 on t2 (c1, c2)`,
		},
		{
			name:      "np name, one column, unique",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
				IsUnique: true,
			},
			expectedStatement: `create unique index idx_t2_c1 on t2 (c1)`,
		},
		{
			name:      "with fillfactor option",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
				With: map[string]string{
					"fillfactor": "70",
				},
			},
			expectedStatement: `create index idx_t2_c1 on t2 (c1) with (fillfactor = 70)`,
		},
		{
			name:      "with multiple options",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
					"c2",
				},
				Name: "idx_custom",
				With: map[string]string{
					"fillfactor":             "80",
					"gin_pending_list_limit": "64",
				},
			},
			expectedStatement: `create index idx_custom on t2 (c1, c2) with (fillfactor = 80, gin_pending_list_limit = 64)`,
		},
		{
			name:      "unique index with with clause",
			tableName: "t2",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
				IsUnique: true,
				With: map[string]string{
					"fillfactor": "90",
				},
			},
			expectedStatement: `create unique index idx_t2_c1 on t2 (c1) with (fillfactor = 90)`,
		},
		{
			name:      "schema-qualified table",
			tableName: "t2",
			schema:    "myschema",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{
					"c1",
				},
				Name: "idx_name",
			},
			expectedStatement: `create index "idx_name" on "myschema"."t2" (c1)`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addIndexStatement := AddIndexStatement(test.schema, test.tableName, test.schemaIndex)

			assert.Equal(t, test.expectedStatement, addIndexStatement)
		})
	}
}

func Test_RemoveIndexStatement(t *testing.T) {
	tests := []struct {
		name              string
		schema            string
		tableName         string
		index             *types.Index
		expectedStatement string
	}{
		{
			name:      "empty schema delegates to unqualified",
			schema:    "",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1"},
			expectedStatement: `drop index "idx_t2_c1"`,
		},
		{
			name:      "public schema delegates to unqualified",
			schema:    "public",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1"},
			expectedStatement: `drop index "idx_t2_c1"`,
		},
		{
			name:      "schema-qualified drop",
			schema:    "myschema",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1"},
			expectedStatement: `drop index "myschema"."idx_t2_c1"`,
		},
		{
			name:      "schema-qualified drop unique",
			schema:    "myschema",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1", IsUnique: true},
			expectedStatement: `drop index if exists "myschema"."idx_t2_c1"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statement := RemoveIndexStatement(test.schema, test.tableName, test.index)
			assert.Equal(t, test.expectedStatement, statement)
		})
	}
}

func Test_RemoveConstraintStatement(t *testing.T) {
	tests := []struct {
		name              string
		schema            string
		tableName         string
		index             *types.Index
		expectedStatement string
	}{
		{
			name:      "empty schema delegates to unqualified",
			schema:    "",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1"},
			expectedStatement: `alter table "t2" drop constraint "idx_t2_c1"`,
		},
		{
			name:      "schema-qualified drop constraint",
			schema:    "myschema",
			tableName: "t2",
			index:     &types.Index{Name: "idx_t2_c1"},
			expectedStatement: `alter table "myschema"."t2" drop constraint "idx_t2_c1"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statement := RemoveConstraintStatement(test.schema, test.tableName, test.index)
			assert.Equal(t, test.expectedStatement, statement)
		})
	}
}
