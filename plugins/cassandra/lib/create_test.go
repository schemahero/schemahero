package cassandra

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateTableStatements(t *testing.T) {
	tests := []struct {
		name               string
		keyspace           string
		tableName          string
		tableSchema        schemasv1alpha4.CassandraTableSchema
		expectedStatements []string
	}{
		{
			name:      "simple",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int)`,
			},
		},
		{
			name:      "with pk",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				PrimaryKey: [][]string{
					{
						"a",
					},
				},
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int, primary key (a))`,
			},
		},
		{
			name:      "with compound pk",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				PrimaryKey: [][]string{
					{
						"a",
					},
					{
						"b",
					},
				},
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
					{
						Name: "b",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int, b int, primary key (a, b))`,
			},
		},
		{
			name:      "with composite partition pk",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				PrimaryKey: [][]string{
					{
						"a", "b",
					},
					{
						"c",
					},
				},
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
					{
						Name: "b",
						Type: "int",
					},
					{
						Name: "c",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int, b int, c int, primary key ((a, b), c))`,
			},
		},
		{
			name:      "clustering order",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				ClusteringOrder: &schemasv1alpha4.CassandraClusteringOrder{
					Column: "a",
				},
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int) with clustering order by (a)`,
			},
		},
		{
			name:      "clustering order desc",
			keyspace:  "k",
			tableName: "t",
			tableSchema: schemasv1alpha4.CassandraTableSchema{
				ClusteringOrder: &schemasv1alpha4.CassandraClusteringOrder{
					Column:       "a",
					IsDescending: &trueValue,
				},
				Columns: []*schemasv1alpha4.CassandraColumn{
					{
						Name: "a",
						Type: "int",
					},
				},
			},
			expectedStatements: []string{
				`create table "k.t" (a int) with clustering order by (a desc)`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			generatedStatements, err := CreateTableStatements(test.keyspace, test.tableName, &test.tableSchema)
			req.NoError(err)
			assert.Equal(t, test.expectedStatements, generatedStatements)
		})
	}
}
