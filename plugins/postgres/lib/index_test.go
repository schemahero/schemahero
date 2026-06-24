package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
)

func Test_AddIndexStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		tableSchema       string
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
			name:        "schema-qualified table, named index",
			tableName:   "alert_events",
			tableSchema: "alerts_events",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{"org_id"},
				Name:    "idx_alert_events_org_id",
			},
			expectedStatement: `create index idx_alert_events_org_id on "alerts_events"."alert_events" (org_id)`,
		},
		{
			name:        "schema-qualified table, auto-generated index name",
			tableName:   "t2",
			tableSchema: "myschema",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{"c1"},
			},
			expectedStatement: `create index idx_t2_c1 on "myschema"."t2" (c1)`,
		},
		{
			name:        "public schema treated as no schema",
			tableName:   "t2",
			tableSchema: "public",
			schemaIndex: &schemasv1alpha4.PostgresqlTableIndex{
				Columns: []string{"c1"},
			},
			expectedStatement: `create index idx_t2_c1 on t2 (c1)`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addIndexStatement := AddIndexStatement(test.tableName, test.tableSchema, test.schemaIndex)

			assert.Equal(t, test.expectedStatement, addIndexStatement)
		})
	}
}
