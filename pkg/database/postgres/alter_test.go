package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ColumnsMatch(t *testing.T) {
	tests := []struct {
		name   string
		col1   types.Column
		col2   types.Column
		expect bool
	}{
		{
			name: "timestamp and timestamp without time zone",
			col1: types.Column{
				Name:     "a",
				DataType: "timestamp",
			},
			col2: types.Column{
				Name:     "a",
				DataType: "timestamp without time zone",
			},
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := columnsMatch(test.col1, test.col2)
			assert.Equal(t, test.expect, actual)
		})
	}
}

func Test_AlterColumnStatments(t *testing.T) {
	defaultEleven := "11"
	defaultEmpty := ""

	tests := []struct {
		name               string
		tableName          string
		desiredColumns     []*schemasv1alpha4.PostgresqlTableColumn
		existingColumn     *types.Column
		expectedStatements []string
	}{
		{
			name:      "no change",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
				},
				{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "integer",
				ColumnDefault: nil,
			},
			expectedStatements: []string{},
		},
		{
			name:      "no change varchar",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "varchar(32)",
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "character varying (32)",
			},
			expectedStatements: []string{},
		},
		{
			name:      "change data type",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
				},
				{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatements: []string{`alter table "t" alter column "b" type integer`},
		},
		{
			name:      "ignore serial",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "serial",
				},
				{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: nil,
			},
			expectedStatements: []string{},
		},
		{
			name:      "drop column",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatements: []string{`alter table "t" drop column "b"`},
		},
		{
			name:      "add not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha4.PostgresqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &falseValue,
				},
			},
			expectedStatements: []string{`alter table "t" alter column "a" set not null`},
		},
		{
			name:      "drop not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha4.PostgresqlTableColumnConstraints{
						NotNull: &falseValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{`alter table "t" alter column "a" drop not null`},
		},
		{
			name:      "no change to not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "t",
					Type: "text",
				},
			},
			existingColumn: &types.Column{
				Name:          "t",
				DataType:      "text",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &falseValue,
				},
			},
			expectedStatements: []string{},
		},
		{
			name:      "no change to not nullable timestamp using short column type",
			tableName: "ts",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "ts",
					Type: "timestamp",
					Constraints: &schemasv1alpha4.PostgresqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "ts",
				DataType:      "timestamp",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{},
		},
		{
			name:      "no change to not nullable timestamp",
			tableName: "ts",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "ts",
					Type: "timestamp with time zone",
					Constraints: &schemasv1alpha4.PostgresqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "ts",
				DataType:      "timestamp with time zone",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{},
		},
		{
			name:      "default set",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name:    "a",
					Type:    "integer",
					Default: &defaultEleven,
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "integer",
			},
			expectedStatements: []string{`alter table "t" alter column "a" set default '11'`},
		},
		{
			name:      "default unset",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: "a",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: &defaultEleven,
			},
			expectedStatements: []string{`alter table "t" alter column "a" drop default`},
		},
		{
			name:      "default empty string",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name:    "a",
					Type:    "varchar (32)",
					Default: &defaultEmpty,
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "character varying (32)",
			},
			expectedStatements: []string{`alter table "t" alter column "a" set default ''`},
		},
		{
			name:      "add null and default",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name:    "a",
					Type:    "varchar (32)",
					Default: &defaultEleven,
					Constraints: &schemasv1alpha4.PostgresqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "character varying (32)",
			},
			expectedStatements: []string{
				`alter table "t" alter column "a" set default '11'`,
				`update "t" set "a"='11' where "a" is null`,
				`alter table "t" alter column "a" set not null`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatements, err := AlterColumnStatements(test.tableName, []string{}, test.desiredColumns, test.existingColumn)
			req.NoError(err)
			assert.Equal(t, test.expectedStatements, generatedStatements)
		})
	}
}
