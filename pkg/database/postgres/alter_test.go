package postgres

import (
	"testing"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AlterColumnStatment(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		desiredColumns    []*schemasv1alpha1.SQLTableColumn
		existingColumn    *types.Column
		expectedStatement string
	}{
		{
			name:      "no change",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "a",
					Type: "integer",
				},
				&schemasv1alpha1.SQLTableColumn{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "integer",
				ColumnDefault: nil,
			},
			expectedStatement: "",
		},
		{
			name:      "change data type",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "a",
					Type: "integer",
				},
				&schemasv1alpha1.SQLTableColumn{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatement: `alter table "t" alter column "b" type integer`,
		},
		{
			name:      "drop column",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "a",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatement: `alter table "t" drop column "b"`,
		},
		{
			name:      "add not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha1.SQLTableColumnConstraints{
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
			expectedStatement: `alter table "t" alter column "a" set not null`,
		},
		{
			name:      "drop not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha1.SQLTableColumnConstraints{
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
			expectedStatement: `alter table "t" alter column "a" drop not null`,
		},
		{
			name:      "no change to not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
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
			expectedStatement: "",
		},
		{
			name:      "no change to not nullable timestamp using short column type",
			tableName: "ts",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "ts",
					Type: "timestamp",
					Constraints: &schemasv1alpha1.SQLTableColumnConstraints{
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
			expectedStatement: "",
		},
		{
			name:      "no change to not nullable timestamp",
			tableName: "ts",
			desiredColumns: []*schemasv1alpha1.SQLTableColumn{
				&schemasv1alpha1.SQLTableColumn{
					Name: "ts",
					Type: "timestamp with time zone",
					Constraints: &schemasv1alpha1.SQLTableColumnConstraints{
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
			expectedStatement: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := AlterColumnStatement(test.tableName, test.desiredColumns, test.existingColumn)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}
