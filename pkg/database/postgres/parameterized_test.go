package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_maybeParseParameterizedColumnType(t *testing.T) {
	tests := []struct {
		name               string
		requestedType      string
		expectedColumnType string
	}{
		{
			name:               "fake",
			requestedType:      "fake",
			expectedColumnType: "",
		},
		{
			name:               "timestamp",
			requestedType:      "timestamp",
			expectedColumnType: "timestamp",
		},
		{
			name:               "timestamp without time zone",
			requestedType:      "timestamp without time zone",
			expectedColumnType: "timestamp without time zone",
		},
		{
			name:               "timestamp (01:02)",
			requestedType:      "timestamp (01:02)",
			expectedColumnType: "timestamp (01:02)",
		},
		{
			name:               "timestamp (01:02) without time zone",
			requestedType:      "timestamp (01:02) without time zone",
			expectedColumnType: "timestamp (01:02) without time zone",
		},
		{
			name:               "character varying (10)",
			requestedType:      "character varying (10)",
			expectedColumnType: "character varying (10)",
		},
		{
			name:               "bit varying (10)",
			requestedType:      "bit varying (10)",
			expectedColumnType: "bit varying (10)",
		},
		{
			name:               "bit varying",
			requestedType:      "bit varying",
			expectedColumnType: "bit varying (1)",
		},
		{
			name:               "bit(10)",
			requestedType:      "bit(10)",
			expectedColumnType: "bit (10)",
		},
		{
			name:               "bit",
			requestedType:      "bit",
			expectedColumnType: "bit (1)",
		},
		{
			name:               "character (10)",
			requestedType:      "character (10)",
			expectedColumnType: "character (10)",
		},
		{
			name:               "character",
			requestedType:      "character",
			expectedColumnType: "character (1)",
		},
		{
			name:               "numeric",
			requestedType:      "numeric",
			expectedColumnType: "numeric",
		},
		{
			name:               "numeric(10)",
			requestedType:      "numeric(10)",
			expectedColumnType: "numeric (10)",
		},
		{
			name:               "numeric  (5,3)",
			requestedType:      "numeric  (5,3)",
			expectedColumnType: "numeric (5, 3)",
		},
		{
			name:               "vector(1533)",
			requestedType:      "vector(1533)",
			expectedColumnType: "vector (1533)",
		},
		{
			name:               "vector ( 3)",
			requestedType:      "vector ( 3)",
			expectedColumnType: "vector (3)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			columnType, err := maybeParseParameterizedColumnType(test.requestedType)
			req := require.New(t)
			req.NoError(err)
			assert.Equal(t, test.expectedColumnType, columnType)
		})
	}
}
