package mysql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_maybeParseParameterizedColumnType(t *testing.T) {
	tests := []struct {
		name                string
		requestedType       string
		expectedColumnType  string
		expectedErrorSubstr string
	}{
		{
			name:               "fake",
			requestedType:      "fake",
			expectedColumnType: "",
		},
		{
			name:               "varchar(10)",
			requestedType:      "varchar(10)",
			expectedColumnType: "varchar (10)",
		},
		{
			name:               "varchar (10)",
			requestedType:      "varchar (10)",
			expectedColumnType: "varchar (10)",
		},
		{
			name:               "varchar",
			requestedType:      "varchar",
			expectedColumnType: "varchar (1)",
		},
		{
			name:               "tinyint",
			requestedType:      "tinyint",
			expectedColumnType: "tinyint (1)",
		},
		{
			name:               "tinyint(1)",
			requestedType:      "tinyint(1)",
			expectedColumnType: "tinyint (1)",
		},
		{
			name:               "bit",
			requestedType:      "bit",
			expectedColumnType: "bit (1)",
		},
		{
			name:               "bit(10)",
			requestedType:      "bit(10)",
			expectedColumnType: "bit (10)",
		},
		{
			name:               "smallint(10)",
			requestedType:      "smallint(10)",
			expectedColumnType: "smallint (10)",
		},
		{
			name:               "smallint",
			requestedType:      "smallint",
			expectedColumnType: "smallint (5)",
		},
		{
			name:               "mediumint(10)",
			requestedType:      "mediumint(10)",
			expectedColumnType: "mediumint (10)",
		},
		{
			name:               "mediumint",
			requestedType:      "mediumint",
			expectedColumnType: "mediumint (9)",
		},
		{
			name:               "int(10)",
			requestedType:      "int(10)",
			expectedColumnType: "int (10)",
		},
		{
			name:               "int",
			requestedType:      "int",
			expectedColumnType: "int (11)",
		},
		{
			name:               "bigint(10)",
			requestedType:      "bigint(10)",
			expectedColumnType: "bigint (10)",
		},
		{
			name:               "bigint",
			requestedType:      "bigint",
			expectedColumnType: "bigint (20)",
		},
		{
			name:               "decimal(65, 1)",
			requestedType:      "decimal(65, 1)",
			expectedColumnType: "decimal (65, 1)",
		},
		{
			name:               "decimal(20)",
			requestedType:      "decimal(20)",
			expectedColumnType: "decimal (20, 0)",
		},
		{
			name:               "decimal",
			requestedType:      "decimal",
			expectedColumnType: "decimal (10, 0)",
		},
		{
			name:               "float (5, 5)",
			requestedType:      "float (5, 5)",
			expectedColumnType: "float (5, 5)",
		},
		{
			name:               "float",
			requestedType:      "float",
			expectedColumnType: "float",
		},
		{
			name:               "float (5)",
			requestedType:      "float (5)",
			expectedColumnType: "float",
		},
		{
			name:               "float (25)",
			requestedType:      "float (25)",
			expectedColumnType: "double",
		},
		{
			name:               "double (10, 10)",
			requestedType:      "double (10, 10)",
			expectedColumnType: "double (10, 10)",
		},
		{
			name:               "double",
			requestedType:      "double",
			expectedColumnType: "double",
		},
		{
			name:               "text(11)",
			requestedType:      "text(11)",
			expectedColumnType: "text (11)",
		},
		{
			name:               "blob(11)",
			requestedType:      "blob(11)",
			expectedColumnType: "blob (11)",
		},
		{
			name:               "blob",
			requestedType:      "blob",
			expectedColumnType: "blob",
		},
		{
			name:               "char (36)",
			requestedType:      "char (36)",
			expectedColumnType: "char (36)",
		},
		{
			name:               "datetime (0)",
			requestedType:      "datetime (0)",
			expectedColumnType: "datetime (0)",
		},
		{
			name:               "datetime (6)",
			requestedType:      "datetime (6)",
			expectedColumnType: "datetime (6)",
		},
		{
			name:                "datetime (7)",
			requestedType:       "datetime (7)",
			expectedColumnType:  "",
			expectedErrorSubstr: "invalid datetime precision 7",
		},
		{
			name:               "timestamp (0)",
			requestedType:      "timestamp (0)",
			expectedColumnType: "timestamp (0)",
		},
		{
			name:               "timestamp (6)",
			requestedType:      "timestamp (6)",
			expectedColumnType: "timestamp (6)",
		},
		{
			name:                "timestamp (7)",
			requestedType:       "timestamp (7)",
			expectedColumnType:  "",
			expectedErrorSubstr: "invalid timestamp precision 7",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			columnType, err := maybeParseParameterizedColumnType(test.requestedType)

			if test.expectedErrorSubstr == "" {
				req.NoError(err)
			} else {
				req.Error(err)
				assert.True(t, strings.Contains(err.Error(), test.expectedErrorSubstr))
			}
			assert.Equal(t, test.expectedColumnType, columnType)
		})
	}
}
