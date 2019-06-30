package postgres

import (
	"testing"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"

	"github.com/stretchr/testify/assert"
)

func Test_AddIndexStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		schemaIndex       *schemasv1alpha2.SQLTableIndex
		expectedStatement string
	}{
		{
			name:      "no name, one column, not specified unique",
			tableName: "t2",
			schemaIndex: &schemasv1alpha2.SQLTableIndex{
				Columns: []string{
					"c1",
				},
			},
			expectedStatement: `create index idx_t2_c1 on t2 (c1)`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addIndexStatement := AddIndexStatement(test.tableName, test.schemaIndex)

			assert.Equal(t, test.expectedStatement, addIndexStatement)
		})
	}
}
