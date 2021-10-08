package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DatabaseNameFromURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "username:password@tcp(host:3306)/dbname?tls=false",
			uri:      "username:password@tcp(host:3306)/dbname?tls=false",
			expected: "dbname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			dbName, err := DatabaseNameFromURI(test.uri)
			req.NoError(err)
			assert.Equal(t, test.expected, dbName)
		})
	}
}

func Test_ensureMultiStatementsTrue(t *testing.T) {
	tests := []struct {
		uri      string
		expected string
		wantErr  bool
	}{
		{
			uri:      "username:password@tcp(host:3306)/dbname?tls=false",
			expected: "username:password@tcp(host:3306)/dbname?multiStatements=true&tls=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got, err := ensureMultiStatementsTrue(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureMultiStatementsTrue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ensureMultiStatementsTrue() = %v, want %v", got, tt.expected)
			}
		})
	}
}
