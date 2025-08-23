package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DatabaseNameFromURI(t *testing.T) {
	tests := []struct {
		name          string
		uri           string
		expected      string
		expectedError bool
	}{
		{
			name:          "postgres://testuser:password@postgresql:5432/testdb?sslmode=disable",
			uri:           "postgres://testuser:password@postgresql:5432/testdb?sslmode=disable",
			expected:      "testdb",
			expectedError: false,
		},
		{
			name:          "postgresql://testuser:password@postgresql:5432/testdb?sslmode=disable",
			uri:           "postgresql://testuser:password@postgresql:5432/testdb?sslmode=disable",
			expected:      "testdb",
			expectedError: false,
		},
		{
			name:          "invalid",
			uri:           "invalid",
			expected:      "",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			dbName, err := DatabaseNameFromURI(test.uri)
			if test.expectedError {
				req.Error(err)
			} else {
				req.NoError(err)
			}
			assert.Equal(t, test.expected, dbName)
		})
	}
}

func Test_parsePostgresVersion(t *testing.T) {
	tests := []struct {
		name            string
		reportedVersion string
		expectedVersion string
	}{
		{
			name:            "9.5",
			reportedVersion: "PostgreSQL 9.5.17 on x86_64-pc-linux-gnu (Debian 9.5.17-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
			expectedVersion: "9.5.17",
		},
		{
			name:            "9.6",
			reportedVersion: "PostgreSQL 9.6.13 on x86_64-pc-linux-gnu (Debian 9.6.13-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
			expectedVersion: "9.6.13",
		},
		{
			name:            "10",
			reportedVersion: "PostgreSQL 10.8 (Debian 10.8-1.pgdg90+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
			expectedVersion: "10.8.0",
		},
		{
			name:            "11",
			reportedVersion: "PostgreSQL 11.3 (Debian 11.3-1.pgdg90+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit",
			expectedVersion: "11.3.0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			engineVersion, err := parsePostgresVersion(test.reportedVersion)
			req.NoError(err)
			assert.Equal(t, test.expectedVersion, engineVersion)
		})
	}
}
