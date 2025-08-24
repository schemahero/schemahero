package cassandra

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trueValue = true
)

func Test_cassandraColumnAsInsert(t *testing.T) {
	tests := []struct {
		name              string
		column            *schemasv1alpha4.CassandraColumn
		expectedStatement string
	}{
		{
			name: "simple",
			column: &schemasv1alpha4.CassandraColumn{
				Name: "c",
				Type: "int",
			},
			expectedStatement: `c int`,
		},
		{
			name: "text",
			column: &schemasv1alpha4.CassandraColumn{
				Name: "t",
				Type: "text",
			},
			expectedStatement: `t text`,
		},
		{
			name: "simple",
			column: &schemasv1alpha4.CassandraColumn{
				Name:     "c",
				Type:     "int",
				IsStatic: &trueValue,
			},
			expectedStatement: `c int static`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := cassandraColumnAsInsert(test.column)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}
