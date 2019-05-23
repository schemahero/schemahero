package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_maybeParseParameterizedColumnType(t *testing.T) {
	none := int64(-1)

	tests := []struct {
		name               string
		requestedType      string
		expectedColumnType string
		expectedMaxLength  int64
	}{
		{
			name:               "fake",
			requestedType:      "fake",
			expectedColumnType: "",
			expectedMaxLength:  none,
		},
		{
			name:               "timestamp",
			requestedType:      "timestamp",
			expectedColumnType: "timestamp",
			expectedMaxLength:  none,
		},
		{
			name:               "timestamp without time zone",
			requestedType:      "timestamp without time zone",
			expectedColumnType: "timestamp without time zone",
			expectedMaxLength:  none,
		},
		{
			name:               "timestamp (01:02)",
			requestedType:      "timestamp (01:02)",
			expectedColumnType: "timestamp (01:02)",
			expectedMaxLength:  none,
		},
		{
			name:               "timestamp (01:02) without time zone",
			requestedType:      "timestamp (01:02) without time zone",
			expectedColumnType: "timestamp (01:02) without time zone",
			expectedMaxLength:  none,
		},
		{
			name:               "character varying (10)",
			requestedType:      "character varying (10)",
			expectedColumnType: "character varying",
			expectedMaxLength:  int64(10),
		},
		{
			name:               "bit varying (10)",
			requestedType:      "bit varying (10)",
			expectedColumnType: "bit varying",
			expectedMaxLength:  int64(10),
		},
		{
			name:               "bit varying",
			requestedType:      "bit varying",
			expectedColumnType: "bit varying",
			expectedMaxLength:  int64(1),
		},
		{
			name:               "bit(10)",
			requestedType:      "bit(10)",
			expectedColumnType: "bit",
			expectedMaxLength:  int64(10),
		},
		{
			name:               "bit",
			requestedType:      "bit",
			expectedColumnType: "bit",
			expectedMaxLength:  int64(1),
		},
		{
			name:               "character (10)",
			requestedType:      "character (10)",
			expectedColumnType: "character",
			expectedMaxLength:  int64(10),
		},
		{
			name:               "character",
			requestedType:      "character",
			expectedColumnType: "character",
			expectedMaxLength:  int64(1),
		},
		{
			name:               "numeric",
			requestedType:      "numeric",
			expectedColumnType: "numeric",
			expectedMaxLength:  none,
		},
		{
			name:               "numeric(10)",
			requestedType:      "numeric(10)",
			expectedColumnType: "numeric (10)",
			expectedMaxLength:  none,
		},
		{
			name:               "numeric  (5,3)",
			requestedType:      "numeric  (5,3)",
			expectedColumnType: "numeric (5, 3)",
			expectedMaxLength:  none,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			columnType, maxLength, err := maybeParseParameterizedColumnType(test.requestedType)
			req := require.New(t)
			req.NoError(err)
			assert.Equal(t, test.expectedColumnType, columnType)

			if test.expectedMaxLength == none {
				assert.Nil(t, maxLength)
			} else {
				assert.Equal(t, test.expectedMaxLength, *maxLength)
			}
		})
	}
}
