package postgres

import (
	"testing"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AlterColumnStatment(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		desiredColumns    []*schemasv1alpha1.PostgresTableColumn
		existingColumn    *Column
		expectedStatement string
	}{
		{
			name:      "no change",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.PostgresTableColumn{
				&schemasv1alpha1.PostgresTableColumn{
					Name: "a",
					Type: "integer",
				},
				&schemasv1alpha1.PostgresTableColumn{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &Column{
				Name:          "b",
				DataType:      "integer",
				ColumnDefault: nil,
			},
			expectedStatement: "",
		},
		{
			name:      "change data type",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.PostgresTableColumn{
				&schemasv1alpha1.PostgresTableColumn{
					Name: "a",
					Type: "integer",
				},
				&schemasv1alpha1.PostgresTableColumn{
					Name: "b",
					Type: "integer",
				},
			},
			existingColumn: &Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatement: `alter table "t" alter column "b" type integer`,
		},
		{
			name:      "drop column",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.PostgresTableColumn{
				&schemasv1alpha1.PostgresTableColumn{
					Name: "a",
					Type: "integer",
				},
			},
			existingColumn: &Column{
				Name:          "b",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
			},
			expectedStatement: `alter table "t" drop column "b"`,
		},
		{
			name:      "add not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.PostgresTableColumn{
				&schemasv1alpha1.PostgresTableColumn{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha1.PostgresTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: nil,
				Constraints: &ColumnConstraints{
					NotNull: &falseValue,
				},
			},
			expectedStatement: `alter table "t" alter column "a" set not null`,
		},
		{
			name:      "drop not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha1.PostgresTableColumn{
				&schemasv1alpha1.PostgresTableColumn{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha1.PostgresTableColumnConstraints{
						NotNull: &falseValue,
					},
				},
			},
			existingColumn: &Column{
				Name:          "a",
				DataType:      "integer",
				ColumnDefault: nil,
				Constraints: &ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: `alter table "t" alter column "a" drop not null`,
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
