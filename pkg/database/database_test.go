package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasMatchingEnv(t *testing.T) {
	tests := []struct {
		name         string
		databaseEnv  string
		possibleEnvs []string
		expected     bool
	}{
		{
			name:         "both empty",
			databaseEnv:  "",
			possibleEnvs: []string{},
			expected:     true,
		},
		{
			name:         "nil database spec",
			databaseEnv:  "",
			possibleEnvs: nil,
			expected:     true,
		},
		{
			name:         "exact match",
			databaseEnv:  "env1",
			possibleEnvs: []string{"env1"},
			expected:     true,
		},
		{
			name:         "exact no match",
			databaseEnv:  "env1",
			possibleEnvs: []string{"env2"},
			expected:     false,
		},
		{
			name:         "not in list",
			databaseEnv:  "env1",
			possibleEnvs: []string{"env2", "env3"},
			expected:     false,
		},
		{
			name:         "in list",
			databaseEnv:  "env1",
			possibleEnvs: []string{"env1", "env2"},
			expected:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &Database{
				Env: test.databaseEnv,
			}
			actual := d.hasMatchingEnv(test.possibleEnvs)
			assert.Equal(t, test.expected, actual)
		})
	}
}
