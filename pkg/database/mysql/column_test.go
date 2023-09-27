package mysql

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mysqlColumnAsInsert(t *testing.T) {
	default11 := "11"
	tests := []struct {
		name              string
		column            *schemasv1alpha4.MysqlTableColumn
		expectedStatement string
	}{
		{
			name: "simple",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "c",
				Type: "integer",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
				Default: &default11,
			},
			expectedStatement: "`c` int (11) not null default '11'",
		},
		{
			name: "auto_increment",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "c",
				Type: "integer",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
				Attributes: &schemasv1alpha4.MysqlTableColumnAttributes{
					AutoIncrement: &trueValue,
				},
			},
			expectedStatement: "`c` int (11) not null auto_increment",
		},
		{
			name: "charset and collation",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name:      "c",
				Type:      "varchar(255)",
				Charset:   "latin1",
				Collation: "latin1_danish_ci",
			},
			expectedStatement: "`c` varchar (255) character set latin1 collate latin1_danish_ci",
		},
		{
			name: "charset and collation not null default",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "c",
				Type: "varchar(255)",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
				Default:   &default11,
				Charset:   "latin1",
				Collation: "latin1_danish_ci",
			},
			expectedStatement: "`c` varchar (255) character set latin1 collate latin1_danish_ci not null default '11'",
		},
		{
			name: "json field type",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "obj",
				Type: "json",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: "`obj` json not null",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatement, err := mysqlColumnAsInsert(test.column)
			req.NoError(err)
			assert.Equal(t, test.expectedStatement, generatedStatement)
		})
	}
}

func Test_InsertColumnStatement(t *testing.T) {
	tests := []struct {
		name              string
		tableName         string
		desiredColumn     *schemasv1alpha4.MysqlTableColumn
		expectedStatement string
	}{
		{
			name:      "add column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "a",
				Type: "integer",
			},
			expectedStatement: "alter table `t` add column `a` int (11)",
		},
		{
			name:      "add not null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: "alter table `t` add column `a` int (11) not null",
		},
		{
			name:      "add null column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "a",
				Type: "integer",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &falseValue,
				},
			},
			expectedStatement: "alter table `t` add column `a` int (11) null",
		},
		{
			name:      "add auto_increment column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "a",
				Type: "integer",
				Attributes: &schemasv1alpha4.MysqlTableColumnAttributes{
					AutoIncrement: &trueValue,
				},
			},
			expectedStatement: "alter table `t` add column `a` int (11) auto_increment",
		},
		{
			name:      "add json column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "a",
				Type: "json",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: "alter table `t` add column `a` json not null",
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

func Test_schemaColumnToMysqlColumn(t *testing.T) {
	tests := []struct {
		name           string
		schemaColumn   *schemasv1alpha4.MysqlTableColumn
		expectedColumn *types.Column
	}{
		{
			name: "varchar (10)",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "vc",
				Type: "varchar (10)",
			},
			expectedColumn: &types.Column{
				Name:          "vc",
				DataType:      "varchar (10)",
				ColumnDefault: nil,
			},
		},
		{
			name: "bool",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "b",
				Type: "bool",
			},
			expectedColumn: &types.Column{
				Name:          "b",
				DataType:      "tinyint (1)",
				ColumnDefault: nil,
			},
		},
		{
			name: "auto_increment",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "c",
				Type: "integer (11)",
				Attributes: &schemasv1alpha4.MysqlTableColumnAttributes{
					AutoIncrement: &trueValue,
				},
			},
			expectedColumn: &types.Column{
				Name:     "c",
				DataType: "int (11)",
				Attributes: &types.ColumnAttributes{
					AutoIncrement: &trueValue,
				},
			},
		},
		{
			name: "json",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "j",
				Type: "json",
			},
			expectedColumn: &types.Column{
				Name:     "j",
				DataType: "json",
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
