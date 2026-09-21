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
			name: "json",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "j",
				Type: "json",
			},
			expectedStatement: "`j` json",
		},
		{
			name: "enum not null",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active','inactive','pending')",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: "`status` enum('active','inactive','pending') not null",
		},
		{
			name: "enum with default",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active','inactive')",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
				Default: func() *string { s := "active"; return &s }(),
			},
			expectedStatement: "`status` enum('active','inactive') not null default 'active'",
		},
		{
			name: "enum nullable",
			column: &schemasv1alpha4.MysqlTableColumn{
				Name: "role",
				Type: "enum('admin','user','guest')",
			},
			expectedStatement: "`role` enum('admin','user','guest')",
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
		{
			name:      "add enum column",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active','inactive')",
				Constraints: &schemasv1alpha4.MysqlTableColumnConstraints{
					NotNull: &trueValue,
				},
			},
			expectedStatement: "alter table `t` add column `status` enum('active','inactive') not null",
		},
		{
			name:      "add enum column with default",
			tableName: "t",
			desiredColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active','inactive')",
				Default: func() *string { s := "active"; return &s }(),
			},
			expectedStatement: "alter table `t` add column `status` enum('active','inactive') default 'active'",
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
		{
			name: "enum('active','inactive')",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active','inactive')",
			},
			expectedColumn: &types.Column{
				Name:     "status",
				DataType: "enum('active','inactive')",
			},
		},
		{
			name: "enum with spaces in type string",
			schemaColumn: &schemasv1alpha4.MysqlTableColumn{
				Name: "status",
				Type: "enum('active', 'inactive', 'pending')",
			},
			expectedColumn: &types.Column{
				Name:     "status",
				DataType: "enum('active','inactive','pending')",
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
