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
		{
			name:                  "tinytext",
			requestedType:         "tinytext",
			expectedUnaliasedType: "tinytext",
		},
		{
			name:                  "tinytext",
			requestedType:         "tinytext (255)",
			expectedUnaliasedType: "tinytext",
		},
		{
			name:                  "mediumtext",
			requestedType:         "mediumtext",
			expectedUnaliasedType: "mediumtext",
		},
		{
			name:                  "mediumtext",
			requestedType:         "mediumtext (16777215)",
			expectedUnaliasedType: "mediumtext",
		},
		{
			name:                  "longtext",
			requestedType:         "longtext",
			expectedUnaliasedType: "longtext",
		},
		{
			name:                  "longtext (4294967295)",
			requestedType:         "longtext (4294967295)",
			expectedUnaliasedType: "longtext",
		},
		{
			name:                  "tinyblob",
			requestedType:         "tinyblob",
			expectedUnaliasedType: "tinyblob",
		},
		{
			name:                  "tinyblob",
			requestedType:         "tinyblob (255)",
			expectedUnaliasedType: "tinyblob",
		},
		{
			name:                  "mediumblob",
			requestedType:         "mediumblob",
			expectedUnaliasedType: "mediumblob",
		},
		{
			name:                  "mediumblob",
			requestedType:         "mediumblob (16777215)",
			expectedUnaliasedType: "mediumblob",
		},
		{
			name:                  "longblob",
			requestedType:         "longblob",
			expectedUnaliasedType: "longblob",
		},
		{
			name:                  "longblob (4294967295)",
			requestedType:         "longblob (4294967295)",
			expectedUnaliasedType: "longblob",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unaliasedType := unaliasUnparameterizedColumnType(test.requestedType)
			assert.Equal(t, test.expectedUnaliasedType, unaliasedType)
		})
	}
}
