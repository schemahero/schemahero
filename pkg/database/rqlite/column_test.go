package rqlite

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_columnAsInsert(t *testing.T) {
	default11 := "11"
	tests := []struct {
		name              string
		column            *schemasv1alpha4.RqliteTableColumn
		expectedStatement string
	}{
		{
			name: "simple",
			column: &schemasv1alpha4.RqliteTableColumn{
				Name: "c",
				Type: "integer",
			},
			expectedStatement: `"c" integer`,
		},
		{
			name: "text",
			column: &schemasv1alpha4.RqliteTableColumn{
				Name: "t",
				Type: "text",
			},
			expectedStatement: `"t" text`,
		},
		{
			name: "constraint not null",
			column: &schemasv1alpha4.RqliteTableColumn{
				Name: "c",
				Type: "integer",
				Constraints: &schemasv1alpha4.RqliteTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: `"c" integer not null`,
		},
		{
			name: "default",
			column: &schemasv1alpha4.RqliteTableColumn{
				Name:    "c",
				Type:    "integer",
				Default: &default11,
			},
			expectedStatement: `"c" integer default '11'`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := rqliteColumnAsInsert(test.column)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}

func Test_InsertColumnStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		desiredColumn     *schemasv1alpha4.RqliteTableColumn
		expectedStatement string
	}{
		{
			name:      "add column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.RqliteTableColumn{
				Name: "a",
				Type: "integer",
			},
			expectedStatement: `alter table "t" add column "a" integer`,
		},
		{
			name:      "add not null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.RqliteTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha4.RqliteTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: `alter table "t" add column "a" integer not null`,
		},
		{
			name:      "add null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.RqliteTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha4.RqliteTableColumnConstraints{
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

func Test_schemaColumnToRqliteColumn(t *testing.T) {
	defaultText := "sometext"
	tests := []struct {
		name           string
		schemaColumn   *schemasv1alpha4.RqliteTableColumn
		expectedColumn *types.Column
	}{
		{
			name: "text",
			schemaColumn: &schemasv1alpha4.RqliteTableColumn{
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
			name: "integer",
			schemaColumn: &schemasv1alpha4.RqliteTableColumn{
				Name: "t",
				Type: "integer",
			},
			expectedColumn: &types.Column{
				Name:          "t",
				DataType:      "integer",
				ColumnDefault: nil,
				Constraints:   nil,
			},
		},
		{
			name: "text default",
			schemaColumn: &schemasv1alpha4.RqliteTableColumn{
				Name:    "t",
				Type:    "text",
				Default: &defaultText,
			},
			expectedColumn: &types.Column{
				Name:          "t",
				DataType:      "text",
				ColumnDefault: &defaultText,
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
