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
