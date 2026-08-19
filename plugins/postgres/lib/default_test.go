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

func Test_normalizePostgresDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "plain literal",
			value:    "pending",
			expected: "pending",
		},
		{
			name:     "quoted literal",
			value:    "'pending'",
			expected: "pending",
		},
		{
			name:     "oid cast",
			value:    "'{}'::jsonb",
			expected: "{}",
		},
		{
			name:     "function unchanged",
			value:    "now()",
			expected: "now()",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, normalizePostgresDefault(test.value))
		})
	}
}

func Test_postgresDefaultsEqual(t *testing.T) {
	quotedPending := "'pending'"
	pending := "pending"
	jsonbCast := "'{}'::jsonb"
	jsonb := "{}"
	now := "now()"

	assert.True(t, postgresDefaultsEqual(&quotedPending, &pending))
	assert.True(t, postgresDefaultsEqual(&jsonbCast, &jsonb))
	assert.True(t, postgresDefaultsEqual(&now, &now))
	assert.False(t, postgresDefaultsEqual(&now, &pending))
	assert.False(t, postgresDefaultsEqual(&now, nil))
	assert.True(t, postgresDefaultsEqual(nil, nil))
}
