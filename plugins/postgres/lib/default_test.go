package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatPostgresDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "integer literal",
			value:    "11",
			expected: "'11'",
		},
		{
			name:     "empty string literal",
			value:    "",
			expected: "''",
		},
		{
			name:     "now function",
			value:    "now()",
			expected: "now()",
		},
		{
			name:     "gen_random_uuid function",
			value:    "gen_random_uuid()",
			expected: "gen_random_uuid()",
		},
		{
			name:     "current_timestamp keyword",
			value:    "CURRENT_TIMESTAMP",
			expected: "CURRENT_TIMESTAMP",
		},
		{
			name:     "already quoted string",
			value:    "'pending'",
			expected: "'pending'",
		},
		{
			name:     "jsonb cast literal",
			value:    "'{}'::jsonb",
			expected: "'{}'::jsonb",
		},
		{
			name:     "true keyword",
			value:    "true",
			expected: "true",
		},
		{
			name:     "plain text still quoted",
			value:    "pending",
			expected: "'pending'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, formatPostgresDefaultValue(test.value))
		})
	}
}
