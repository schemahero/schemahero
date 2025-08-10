package datamigration

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUpdateSQL(t *testing.T) {
	tests := []struct {
		name      string
		dbType    string
		migration schemasv1alpha4.DataMigrationOperation
		expected  string
		wantErr   bool
	}{
		{
			name:   "simple update",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:   schemasv1alpha4.UpdateMigration,
				Table:  "users",
				Column: "status",
				Value:  "'active'",
			},
			expected: "UPDATE users SET status = 'active'",
			wantErr:  false,
		},
		{
			name:   "update with where clause",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:   schemasv1alpha4.UpdateMigration,
				Table:  "users",
				Column: "status",
				Value:  "'active'",
				Where:  "status IS NULL",
			},
			expected: "UPDATE users SET status = 'active' WHERE status IS NULL",
			wantErr:  false,
		},
		{
			name:   "missing required fields",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:  schemasv1alpha4.UpdateMigration,
				Table: "users",
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := generateUpdateSQL(tt.dbType, tt.migration)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, sql)
			}
		})
	}
}

func TestGenerateCalculateSQL(t *testing.T) {
	tests := []struct {
		name      string
		dbType    string
		migration schemasv1alpha4.DataMigrationOperation
		expected  string
		wantErr   bool
	}{
		{
			name:   "calculate full name",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:       schemasv1alpha4.CalculateMigration,
				Table:      "users",
				Column:     "full_name",
				Expression: "first_name || ' ' || last_name",
			},
			expected: "UPDATE users SET full_name = first_name || ' ' || last_name",
			wantErr:  false,
		},
		{
			name:   "calculate with where clause",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:       schemasv1alpha4.CalculateMigration,
				Table:      "users",
				Column:     "display_name",
				Expression: "UPPER(username)",
				Where:      "display_name IS NULL",
			},
			expected: "UPDATE users SET display_name = UPPER(username) WHERE display_name IS NULL",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := generateCalculateSQL(tt.dbType, tt.migration)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, sql)
			}
		})
	}
}

func TestGenerateConvertSQL(t *testing.T) {
	tests := []struct {
		name      string
		dbType    string
		migration schemasv1alpha4.DataMigrationOperation
		expected  string
		wantErr   bool
	}{
		{
			name:   "postgres timestamp to timestamptz",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:   schemasv1alpha4.ConvertMigration,
				Table:  "events",
				Column: "created_at",
				From:   "timestamp",
				To:     "timestamptz",
			},
			expected: "ALTER TABLE events ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC'",
			wantErr:  false,
		},
		{
			name:   "mysql timestamp conversion",
			dbType: "mysql",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:   schemasv1alpha4.ConvertMigration,
				Table:  "events",
				Column: "created_at",
				From:   "timestamp",
				To:     "timestamptz",
			},
			expected: "ALTER TABLE events MODIFY COLUMN created_at DATETIME",
			wantErr:  false,
		},
		{
			name:   "generic type conversion",
			dbType: "postgres",
			migration: schemasv1alpha4.DataMigrationOperation{
				Type:   schemasv1alpha4.ConvertMigration,
				Table:  "users",
				Column: "age",
				From:   "varchar",
				To:     "integer",
			},
			expected: "ALTER TABLE users ALTER COLUMN age TYPE integer",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := generateConvertSQL(tt.dbType, tt.migration)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, sql)
			}
		})
	}
}