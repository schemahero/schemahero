package sqlite

import (
	"testing"

	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
)

func Test_ColumnsMatch(t *testing.T) {
	tests := []struct {
		name   string
		col1   types.Column
		col2   types.Column
		expect bool
	}{
		{
			name: "integers",
			col1: types.Column{
				Name:     "a",
				DataType: "integer",
			},
			col2: types.Column{
				Name:     "a",
				DataType: "integer",
			},
			expect: true,
		},
		{
			name: "reals",
			col1: types.Column{
				Name:     "a",
				DataType: "real",
			},
			col2: types.Column{
				Name:     "a",
				DataType: "real",
			},
			expect: true,
		},
		{
			name: "string and integer",
			col1: types.Column{
				Name:     "a",
				DataType: "string",
			},
			col2: types.Column{
				Name:     "a",
				DataType: "integer",
			},
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := columnsMatch(test.col1, test.col2)
			assert.Equal(t, test.expect, actual)
		})
	}
}
