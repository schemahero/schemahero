package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_configMapNameForMigration(t *testing.T) {
	tests := []struct {
		name         string
		databaseName string
		tableName    string
		migrationID  string
		expect       string
	}{
		{
			name:         "short-enough",
			databaseName: "a",
			tableName:    "b",
			migrationID:  "c",
			expect:       "a-b-c",
		},
		{
			name:         "should-be-table-and-id",
			databaseName: "a-long-database-name-a-long-database-name-a-long-database-name-a-long-database-name-a-long-database-name-a-long-database-name",
			tableName:    "a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name",
			migrationID:  "a-migration-id",
			expect:       "a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-long-table-name-a-migration-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := configMapNameForMigration(test.databaseName, test.tableName, test.migrationID)
			assert.Equal(t, test.expect, actual)
		})
	}
}
