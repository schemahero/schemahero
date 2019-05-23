package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_unaliasParameterizedColumnType(t *testing.T) {
	aliasedParmeterizedTests := map[string]string{
		"char      (1)":             "character (1)",
		"integer(1)":                "int (1)",
		"dec(5,5)":                  "decimal (5, 5)",
		"dec":                       "decimal (10, 0)",
		"dec (5)":                   "decimal (5, 0)",
		"double precision":          "double",
		"double precision (10, 10)": "double (10, 10)",
	}

	for aliased, expectedUnaliased := range aliasedParmeterizedTests {
		t.Run(aliased, func(t *testing.T) {
			unaliasedType := unaliasParameterizedColumnType(aliased)
			assert.Equal(t, expectedUnaliased, unaliasedType)
		})
	}
}

func Test_unaliasUnparameterizedColumnType(t *testing.T) {
	tests := []struct {
		name                  string
		requestedType         string
		expectedUnaliasedType string
	}{
		{
			name:                  "bool",
			requestedType:         "bool",
			expectedUnaliasedType: "tinyint (1)",
		},
		{
			name:                  "boolean",
			requestedType:         "boolean",
			expectedUnaliasedType: "tinyint (1)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unaliasedType := unaliasUnparameterizedColumnType(test.requestedType)
			assert.Equal(t, test.expectedUnaliasedType, unaliasedType)
		})
	}
}
