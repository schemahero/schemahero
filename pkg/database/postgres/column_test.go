package postgres

import (
	"testing"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_postgresColumnAsInsert(t *testing.T) {
	default11 := "11"
	tests := []struct {
		name              string
		column            *schemasv1alpha3.SQLTableColumn
		expectedStatement string
	}{
		{
			name: "simple",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "c",
				Type: "integer",
			},
			expectedStatement: `"c" integer`,
		},
		{
			name: "text",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "t",
				Type: "text",
			},
			expectedStatement: `"t" text`,
		},
		{
			name: "timestamp without time zone",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "t",
				Type: "timestamp without time zone",
			},
			expectedStatement: `"t" timestamp without time zone`,
		},
		{
			name: "character varying (4)",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "c",
				Type: "character varying (4)",
			},
			expectedStatement: `"c" character varying (4)`,
		},
		{
			name: "constraint not null",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "c",
				Type: "integer",
				Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: `"c" integer not null`,
		},
		{
			name: "default",
			column: &schemasv1alpha3.SQLTableColumn{
				Name:    "c",
				Type:    "integer",
				Default: &default11,
			},
			expectedStatement: `"c" integer default '11'`,
		},
		{
			name: "text[]",
			column: &schemasv1alpha3.SQLTableColumn{
				Name: "c",
				Type: "text[]",
			},
			expectedStatement: `"c" text[]`,
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
		desiredColumn     *schemasv1alpha3.SQLTableColumn
		expectedStatement string
	}{
		{
			name:      "add column",
			tableName: "t",
			desiredColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "a",
				Type: "integer",
			},
			expectedStatement: `alter table "t" add column "a" integer`,
		},
		{
			name:      "add not null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: `alter table "t" add column "a" integer not null`,
		},
		{
			name:      "add null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha3.SQLTableColumnConstraints{
					NotNull: &falseValue,
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
		schemaColumn   *schemasv1alpha3.SQLTableColumn
		expectedColumn *types.Column
	}{
		{
			name: "text",
			schemaColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "t",
				Type: "text",
			},
			expectedColumn: &types.Column{
				Name:          "t",
				DataType:      "text",
				ColumnDefault: nil,
				Constraints:   nil,
			},
		},
		{
			name: "character varying (10)",
			schemaColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "c",
				Type: "character varying (10)",
			},
			expectedColumn: &types.Column{
				Name:          "c",
				DataType:      "character varying (10)",
				ColumnDefault: nil,
			},
		},
		{
			name: "varchar (10)",
			schemaColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "vc",
				Type: "varchar (10)",
			},
			expectedColumn: &types.Column{
				Name:          "vc",
				DataType:      "character varying (10)",
				ColumnDefault: nil,
			},
		},
		{
			name: "cidr",
			schemaColumn: &schemasv1alpha3.SQLTableColumn{
				Name: "ip",
				Type: "cidr",
			},
			expectedColumn: &types.Column{
				Name:          "ip",
				DataType:      "cidr",
				ColumnDefault: nil,
				Constraints:   nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			column, err := schemaColumnToColumn(test.schemaColumn)
			req.NoError(err)
			assert.Equal(t, test.expectedColumn, column)
		})
	}

}
