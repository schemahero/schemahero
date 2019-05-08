package postgres

import (
	"testing"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_unaliasParameterizedColumnType(t *testing.T) {
	tests := []struct {
		name                  string
		requestedType         string
		expectedUnaliasedType string
	}{
		{
			name:                  "varchar(255)",
			requestedType:         "varchar(255)",
			expectedUnaliasedType: "character varying (255)",
		}, {
			name:                  "varchar",
			requestedType:         "varchar",
			expectedUnaliasedType: "character varying",
		}, {
			name:                  "varchar (100)",
			requestedType:         "varchar (100)",
			expectedUnaliasedType: "character varying (100)",
		}, {
			name:                  "varbit (50)",
			requestedType:         "varbit (50)",
			expectedUnaliasedType: "bit varying (50)",
		}, {
			name:                  "char",
			requestedType:         "char",
			expectedUnaliasedType: "character",
		}, {
			name:                  "char(36)",
			requestedType:         "char(36)",
			expectedUnaliasedType: "character (36)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unaliasedType := unaliasParameterizedColumnType(test.requestedType)
			assert.Equal(t, test.expectedUnaliasedType, unaliasedType)
		})
	}
}

func Test_unaliasSimpleColumnType(t *testing.T) {
	tests := []struct {
		name                  string
		requestedType         string
		expectedUnaliasedType string
	}{
		{
			name:                  "int8",
			requestedType:         "int8",
			expectedUnaliasedType: "bigint",
		},
		{
			name:                  "serial8",
			requestedType:         "serial8",
			expectedUnaliasedType: "bigserial",
		},
		{
			name:                  "bool",
			requestedType:         "bool",
			expectedUnaliasedType: "boolean",
		},
		{
			name:                  "int",
			requestedType:         "int",
			expectedUnaliasedType: "integer",
		},
		{
			name:                  "int4",
			requestedType:         "int4",
			expectedUnaliasedType: "integer",
		},
		{
			name:                  "float4",
			requestedType:         "float4",
			expectedUnaliasedType: "real",
		},
		{
			name:                  "int2",
			requestedType:         "int2",
			expectedUnaliasedType: "smallint",
		},
		{
			name:                  "serial2",
			requestedType:         "serial2",
			expectedUnaliasedType: "smallserial",
		},
		{
			name:                  "serial4",
			requestedType:         "serial4",
			expectedUnaliasedType: "serial",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unaliasedType := unaliasSimpleColumnType(test.requestedType)
			assert.Equal(t, test.expectedUnaliasedType, unaliasedType)
		})
	}
}

func Test_postgresColumnAsInsert(t *testing.T) {
	tests := []struct {
		name              string
		column            *schemasv1alpha1.PostgresTableColumn
		expectedStatement string
	}{
		{
			name: "simple",
			column: &schemasv1alpha1.PostgresTableColumn{
				Name: "c",
				Type: "integer",
			},
			expectedStatement: `"c" integer`,
		},
		// {
		// 	name: "needs_escape",
		// 	column: &schemasv1alpha1.PostgresTableColumn{
		// 		Name: "year",
		// 		Type: "fake",
		// 	},
		// 	expectedStatement: `"year" "fake"`,
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := postgresColumnAsInsert(test.column)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}

func Test_InsertColumnStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		desiredColumn     *schemasv1alpha1.PostgresTableColumn
		expectedStatement string
	}{
		{
			name:      "add column",
			tableName: "t",
			desiredColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "a",
				Type: "integer",
			},
			expectedStatement: `alter table "t" add column "a" integer`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := InsertColumnStatement(test.tableName, test.desiredColumn)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}

// func Test_columnTypeToPostgresColumn(t *testing.T) {
// 	// translated := translatePostgresColumnType("integer")
// 	// assert.Equal(t, "bigint", translated, "integer should translate to bigint")
// }
