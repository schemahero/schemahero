package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateStaticUpdateStatement(t *testing.T) {
	tests := []struct {
		name      string
		operation schemasv1alpha4.StaticUpdateOperation
		expected  string
		expectErr bool
	}{
		{
			name: "simple static update",
			operation: schemasv1alpha4.StaticUpdateOperation{
				Table: "users",
				Set: map[string]string{
					"status": "active",
					"updated_at": "CURRENT_TIMESTAMP",
				},
			},
			expected: `UPDATE "users" SET "status" = 'active', "updated_at" = CURRENT_TIMESTAMP`,
		},
		{
			name: "static update with where clause",
			operation: schemasv1alpha4.StaticUpdateOperation{
				Table: "users",
				Set: map[string]string{
					"status": "inactive",
				},
				Where: "last_login < CURRENT_DATE - INTERVAL '1 year'",
			},
			expected: `UPDATE "users" SET "status" = 'inactive' WHERE last_login < CURRENT_DATE - INTERVAL '1 year'`,
		},
		{
			name: "empty table name",
			operation: schemasv1alpha4.StaticUpdateOperation{
				Set: map[string]string{"status": "active"},
			},
			expectErr: true,
		},
		{
			name: "empty set clause",
			operation: schemasv1alpha4.StaticUpdateOperation{
				Table: "users",
				Set:   map[string]string{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateStaticUpdateStatement(tt.operation)
			
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			// Since map iteration order is not guaranteed, we need to check both possible orders
			assert.Contains(t, []string{tt.expected, 
				`UPDATE "users" SET "updated_at" = CURRENT_TIMESTAMP, "status" = 'active'`}, result)
		})
	}
}

func TestGenerateCalculatedUpdateStatement(t *testing.T) {
	tests := []struct {
		name      string
		operation schemasv1alpha4.CalculatedUpdateOperation
		expected  string
		expectErr bool
	}{
		{
			name: "simple calculated update",
			operation: schemasv1alpha4.CalculatedUpdateOperation{
				Table: "users",
				Calculations: []schemasv1alpha4.ColumnCalculation{
					{
						Column:     "full_name",
						Expression: "CONCAT(first_name, ' ', last_name)",
					},
				},
			},
			expected: `UPDATE "users" SET "full_name" = CONCAT(first_name, ' ', last_name)`,
		},
		{
			name: "multiple calculations with where",
			operation: schemasv1alpha4.CalculatedUpdateOperation{
				Table: "orders",
				Calculations: []schemasv1alpha4.ColumnCalculation{
					{
						Column:     "total_with_tax",
						Expression: "subtotal * 1.08",
					},
					{
						Column:     "discount_amount",
						Expression: "subtotal * discount_percent / 100",
					},
				},
				Where: "total_with_tax IS NULL",
			},
			expected: `UPDATE "orders" SET "total_with_tax" = subtotal * 1.08, "discount_amount" = subtotal * discount_percent / 100 WHERE total_with_tax IS NULL`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateCalculatedUpdateStatement(tt.operation)
			
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateTimezoneConversionExpression(t *testing.T) {
	tests := []struct {
		name      string
		transform schemasv1alpha4.DataTransformation
		expected  string
		expectErr bool
	}{
		{
			name: "timezone conversion",
			transform: schemasv1alpha4.DataTransformation{
				Column:    "created_at",
				FromValue: "UTC",
				ToValue:   "America/New_York",
			},
			expected: `"created_at" AT TIME ZONE 'UTC' AT TIME ZONE 'America/New_York'`,
		},
		{
			name: "missing from timezone",
			transform: schemasv1alpha4.DataTransformation{
				Column:  "created_at",
				ToValue: "America/New_York",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateTimezoneConversionExpression(tt.transform)
			
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCustomSQL(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		expectErr bool
	}{
		{
			name: "valid update statement",
			sql:  "UPDATE users SET status = 'active'",
		},
		{
			name: "valid insert statement",
			sql:  "INSERT INTO logs (message) VALUES ('test')",
		},
		{
			name: "dangerous drop table",
			sql:  "DROP TABLE users",
			expectErr: true,
		},
		{
			name: "dangerous delete",
			sql:  "DELETE FROM users",
			expectErr: true,
		},
		{
			name: "invalid select statement",
			sql:  "SELECT * FROM users",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomSQL(tt.sql)
			
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuoteSQLValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "simple string",
			value:    "active",
			expected: "'active'",
		},
		{
			name:     "string with single quote",
			value:    "don't",
			expected: "'don''t'",
		},
		{
			name:     "SQL function",
			value:    "CURRENT_TIMESTAMP",
			expected: "CURRENT_TIMESTAMP",
		},
		{
			name:     "NOW function",
			value:    "NOW()",
			expected: "NOW()",
		},
		{
			name:     "NULL value",
			value:    "NULL",
			expected: "NULL",
		},
		{
			name:     "Expression with parentheses",
			value:    "CONCAT(first_name, ' ', last_name)",
			expected: "CONCAT(first_name, ' ', last_name)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteSQLValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
