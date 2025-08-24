package mysql

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatement(t *testing.T) {
	tests := []struct {
		name               string
		tableSchema        *schemasv1alpha4.MysqlTableSchema
		tableName          string
		expectedStatements []string
	}{
		{
			name: "simple",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
			},
			tableName: "simple",
			expectedStatements: []string{
				"create table `simple` (`id` int (11), primary key (`id`))",
			},
		},
		{
			name: "varchar composite primary key",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"col_a",
					"col_b",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "col_a",
						Type: "char (36)",
					},
					{
						Name: "col_b",
						Type: "varchar (255)",
					},
				},
			},
			tableName: "table_b",
			expectedStatements: []string{
				"create table `table_b` (`col_a` char (36), `col_b` varchar (255), primary key (`col_a`, `col_b`))",
			},
		},
		{
			name: "composite primary key",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"one",
					"two",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "one",
						Type: "integer",
					},
					{
						Name: "two",
						Type: "integer",
					},
					{
						Name: "three",
						Type: "varchar(255)",
					},
				},
			},
			tableName: "composite_primary_key",
			expectedStatements: []string{
				"create table `composite_primary_key` (`one` int (11), `two` int (11), `three` varchar (255), primary key (`one`, `two`))",
			},
		},
		{
			name: "decimal (8, 2) column",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"one",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "one",
						Type: "integer",
					},
					{
						Name: "bee",
						Type: "decimal (8, 2)",
					},
				},
			},
			tableName: "decimal_8_2",
			expectedStatements: []string{
				"create table `decimal_8_2` (`one` int (11), `bee` decimal (8, 2), primary key (`one`))",
			},
		},
		{
			name: "table with default charset and collate",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
				DefaultCharset: "latin1",
				Collation:      "latin1_german1_ci",
			},
			tableName: "test",
			expectedStatements: []string{
				"create table `test` (`id` int (11), primary key (`id`)) default character set latin1 collate latin1_german1_ci",
			},
		},
		{
			name: "table with default charset without collate",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
				DefaultCharset: "latin1",
			},
			tableName: "test",
			expectedStatements: []string{
				"create table `test` (`id` int (11), primary key (`id`)) default character set latin1",
			},
		},
		{
			name: "table with collate and no default character set",
			tableSchema: &schemasv1alpha4.MysqlTableSchema{
				PrimaryKey: []string{
					"id",
				},
				Columns: []*schemasv1alpha4.MysqlTableColumn{
					{
						Name: "id",
						Type: "integer",
					},
				},
				Collation: "latin1_german1_ci",
			},
			tableName: "test",
			expectedStatements: []string{
				"create table `test` (`id` int (11), primary key (`id`)) collate latin1_german1_ci",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			createTableStatements, err := CreateTableStatements(test.tableName, test.tableSchema)
			req.NoError(err)

			assert.Equal(t, test.expectedStatements, createTableStatements)
		})
	}
}
