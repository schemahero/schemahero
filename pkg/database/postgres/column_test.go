package postgres

import (
	"testing"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ten = int64(10)
)

func Test_unaliasParameterizedColumnType(t *testing.T) {
	aliasedParmeterizedTests := map[string]string{
		"varchar(255)":                       "character varying (255)",
		"varchar      (1)":                   "character varying (1)",
		"varchar ( 10 )":                     "character varying (10)",
		"varchar":                            "character varying",
		"varchar (100)":                      "character varying (100)",
		"varbit (50)":                        "bit varying (50)",
		"char":                               "character",
		"char(36)":                           "character (36)",
		"decimal":                            "numeric",
		"decimal (10)":                       "numeric (10)",
		"decimal (10, 5)":                    "numeric (10, 5)",
		"decimal (10,5)":                     "numeric (10, 5)",
		"decimal(10, 5)":                     "numeric (10, 5)",
		"decimal(10,5)":                      "numeric (10, 5)",
		"decimal(   10,    5 )":              "numeric (10, 5)",
		"timetz":                             "time with time zone",
		"timetz(01:02)":                      "time (01:02) with time zone",
		"timetz (2006-01-02T15:04:05Z07:00)": "time (2006-01-02T15:04:05Z07:00) with time zone",
		"timestamptz":                        "timestamp with time zone",
		"timestamptz(01:02)":                 "timestamp (01:02) with time zone",
		"timestamptz (2006-01-02T15:04:05Z07:00)": "timestamp (2006-01-02T15:04:05Z07:00) with time zone",
	}

	for aliased, expectedUnaliased := range aliasedParmeterizedTests {
		t.Run(aliased, func(t *testing.T) {
			unaliasedType := unaliasParameterizedColumnType(aliased)
			assert.Equal(t, expectedUnaliased, unaliasedType)
		})
	}
}

func Test_maybeParseParameterizedColumnType(t *testing.T) {
	parameterizedTests := map[string]string{
		"fake":                        "",
		"timestamp":                   "timestamp",
		"timestamp without time zone": "timestamp without time zone",
		// "timestamp (01:02)":                   "timestamp (01:02)",
		// "timestamp (01:02) without time zone": "timestamp (01:02) without time zone",
	}

	for input, expectedOutput := range parameterizedTests {
		t.Run(input, func(t *testing.T) {
			output, _, _ := maybeParseParameterizedColumnType(input)
			assert.Equal(t, expectedOutput, output)
		})
	}
}

func Test_unaliasUnparameterizedColumnType(t *testing.T) {
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
		{
			name:                  "cidr",
			requestedType:         "cidr",
			expectedUnaliasedType: "cidr",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unaliasedType := unaliasUnparameterizedColumnType(test.requestedType)
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
		{
			name: "text",
			column: &schemasv1alpha1.PostgresTableColumn{
				Name: "t",
				Type: "text",
			},
			expectedStatement: `"t" text`,
		},
		{
			name: "timestamp without time zone",
			column: &schemasv1alpha1.PostgresTableColumn{
				Name: "t",
				Type: "timestamp without time zone",
			},
			expectedStatement: `"t" timestamp without time zone`,
		},
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
		{
			name:      "add not null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha1.PostgresTableColumnConstraints{
					NotNull: true,
				},
			},
			expectedStatement: `alter table "t" add column "a" integer not null`,
		},
		{
			name:      "add null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha1.PostgresTableColumnConstraints{
					NotNull: false,
				},
			},
			expectedStatement: `alter table "t" add column "a" integer null`,
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

func Test_schemaColumnToPostgresColumn(t *testing.T) {
	tests := []struct {
		name           string
		schemaColumn   *schemasv1alpha1.PostgresTableColumn
		expectedColumn *Column
	}{
		{
			name: "text",
			schemaColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "t",
				Type: "text",
			},
			expectedColumn: &Column{
				DataType:      "text",
				ColumnDefault: nil,
				Constraints:   nil,
			},
		},
		{
			name: "character varying (10)",
			schemaColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "c",
				Type: "character varying (10)",
			},
			expectedColumn: &Column{
				DataType:      "character varying",
				ColumnDefault: nil,
				Constraints: &ColumnConstraints{
					MaxLength: &ten,
				},
			},
		},
		{
			name: "cidr",
			schemaColumn: &schemasv1alpha1.PostgresTableColumn{
				Name: "ip",
				Type: "cidr",
			},
			expectedColumn: &Column{
				DataType:      "cidr",
				ColumnDefault: nil,
				Constraints:   nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			column, err := schemaColumnToPostgresColumn(test.schemaColumn)
			req.NoError(err)
			assert.Equal(t, test.expectedColumn, column)
		})
	}

}
