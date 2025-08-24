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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addIndexStatement := AddIndexStatement(test.tableName, test.schemaIndex)

			assert.Equal(t, test.expectedStatement, addIndexStatement)
		})
	}
}
