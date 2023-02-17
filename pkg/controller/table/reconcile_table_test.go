package table

import (
	"testing"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
)

func Test_checkDatabaseTypeMatches(t *testing.T) {
	tests := []struct {
		name        string
		connection  databasesv1alpha4.DatabaseConnection
		tableSchema schemasv1alpha4.TableSchema
		expect      bool
	}{
		{
			name: "is postgres",
			connection: databasesv1alpha4.DatabaseConnection{
				Postgres: &databasesv1alpha4.PostgresConnection{
					URI: databasesv1alpha4.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				Postgres: &schemasv1alpha4.PostgresqlTableSchema{},
			},
			expect: true,
		},
		{
			name: "is mysql",
			connection: databasesv1alpha4.DatabaseConnection{
				Mysql: &databasesv1alpha4.MysqlConnection{
					URI: databasesv1alpha4.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				Mysql: &schemasv1alpha4.MysqlTableSchema{},
			},
			expect: true,
		},
		{
			name: "is mysql, expect postgres",
			connection: databasesv1alpha4.DatabaseConnection{
				Mysql: &databasesv1alpha4.MysqlConnection{
					URI: databasesv1alpha4.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				Postgres: &schemasv1alpha4.PostgresqlTableSchema{},
			},
			expect: false,
		},
		{
			name: "is postgres, expect mysql",
			connection: databasesv1alpha4.DatabaseConnection{
				Postgres: &databasesv1alpha4.PostgresConnection{
					URI: databasesv1alpha4.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				Mysql: &schemasv1alpha4.MysqlTableSchema{},
			},
			expect: false,
		},
		{
			name: "is rqlite",
			connection: databasesv1alpha4.DatabaseConnection{
				RQLite: &databasesv1alpha4.RqliteConnection{
					URI: databasesv1alpha4.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				RQLite: &schemasv1alpha4.RqliteTableSchema{},
			},
			expect: true,
		},
		{
			name: "is cassandra",
			connection: databasesv1alpha4.DatabaseConnection{
				Cassandra: &databasesv1alpha4.CassandraConnection{
					Hosts: []string{"test"},
				},
			},
			tableSchema: schemasv1alpha4.TableSchema{
				Cassandra: &schemasv1alpha4.CassandraTableSchema{},
			},
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := checkDatabaseTypeMatches(&test.connection, &test.tableSchema)
			assert.Equal(t, test.expect, actual)
		})
	}
}
