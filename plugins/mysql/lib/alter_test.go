package mysql

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AlterColumnStatment(t *testing.T) {
	defaultEleven := "11"
	defaultEmpty := ""

	tests := []struct {
		name               string
		tableName          string
		desiredColumns     []*schemasv1alpha4.MysqlTableColumn
		existingColumn     *types.Column
		expectedStatements []string
		defaultCharset     string
		defaultCollation   string
	}{
		{
			name:      "no change",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
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
				DataType:      "int (11)",
				ColumnDefault: nil,
			},
			expectedStatements: []string{},
		},
		{
			name:      "change data type",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
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
			expectedStatements: []string{"alter table `t` modify column `b` int (11)"},
		},
		{
			name:      "drop column",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "a",
					Type: "integer",
				},
			},
			existingColumn: &types.Column{
				Name:          "b",
				DataType:      "varchar (255)",
				ColumnDefault: nil,
			},
			expectedStatements: []string{"alter table `t` drop column `b`"},
		},
		{
			name:      "add not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "int (11)",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &falseValue,
				},
			},
			expectedStatements: []string{"alter table `t` modify column `a` int (11) not null"},
		},
		{
			name:      "drop not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &falseValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "int (11)",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{"alter table `t` modify column `a` int (11) null"},
		},
		{
			name:      "no change to not null constraint",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
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
			name:      "type change, constraint no change",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "a",
					Type: "integer",
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:          "a",
				DataType:      "varchar(255)",
				ColumnDefault: nil,
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{"alter table `t` modify column `a` int (11) not null"},
		},
		{
			name:      "default set",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
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
			expectedStatements: []string{"alter table `t` modify column `a` int (11) default \"11\""},
		},
		{
			name:      "default unset",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
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
			expectedStatements: []string{"alter table `t` modify column `a` int (11)"},
		},
		{
			name:      "default empty string",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name:    "a",
					Type:    "varchar (32)",
					Default: &defaultEmpty,
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "varchar (32)",
			},
			expectedStatements: []string{"alter table `t` modify column `a` varchar (32) default \"\""},
		},

		// {
		// 	name:      "no change to not nullable timestamp using short column type",
		// 	tableName: "ts",
		// 	desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
		// 		&schemasv1alpha4.MysqlTableColumn{
		// 			Name: "ts",
		// 			Type: "timestamp",
		// 			Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
		// 				NotNull: &trueValue,
		// 			},
		// 		},
		// 	},
		// 	existingColumn: &Column{
		// 		Name:          "ts",
		// 		DataType:      "timestamp",
		// 		ColumnDefault: nil,
		// 		Constraints: &ColumnConstraints{
		// 			NotNull: &trueValue,
		// 		},
		// 	},
		// 	expectedStatement: "",
		// },
		// {
		// 	name:      "no change to not nullable timestamp",
		// 	tableName: "ts",
		// 	desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
		// 		&schemasv1alpha4.MysqlTableColumn{
		// 			Name: "ts",
		// 			Type: "timestamp with time zone",
		// 			Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
		// 				NotNull: &trueValue,
		// 			},
		// 		},
		// 	},
		// 	existingColumn: &Column{
		// 		Name:          "ts",
		// 		DataType:      "timestamp",
		// 		ColumnDefault: nil,
		// 		Constraints: &ColumnConstraints{
		// 			NotNull: &trueValue,
		// 		},
		// 	},
		// 	expectedStatement: "",
		// },
		{
			name:      "add null and default",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name:    "a",
					Type:    "varchar (255)",
					Default: &defaultEleven,
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:     "a",
				DataType: "varchar (255)",
			},
			expectedStatements: []string{
				"alter table `t` modify column `a` varchar (255) default \"11\"",
				"update `t` set `a`=\"11\" where `a` is null",
				"alter table `t` modify column `a` varchar (255) not null default \"11\"",
			},
		},
		{
			name:      "add auto_increment attibute",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "c",
					Type: "integer",
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &trueValue,
					},
					Attributes: &schemasv1alpha4.MysqlTableColumnAttributes{
						AutoIncrement: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:     "c",
				DataType: "integer",
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatements: []string{
				"alter table `t` modify column `c` int (11) not null auto_increment",
			},
		},
		{
			name:      "drop auto_increment attibute",
			tableName: "t",
			desiredColumns: []*schemasv1alpha4.MysqlTableColumn{
				{
					Name: "c",
					Type: "integer",
					Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			existingColumn: &types.Column{
				Name:     "c",
				DataType: "integer",
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
				Attributes: &types.ColumnAttributes{
					AutoIncrement: &trueValue,
				},
			},
			expectedStatements: []string{
				"alter table `t` modify column `c` int (11) not null",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatements, err := AlterColumnStatements(test.tableName, []string{}, test.desiredColumns, test.existingColumn, test.defaultCharset, test.defaultCollation)
			req.NoError(err)
			assert.Equal(t, test.expectedStatements, generatedStatements)
		})
	}
}
