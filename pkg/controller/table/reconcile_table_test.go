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
				Postgres: &schemasv1alpha4.SQLTableSchema{},
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
				Mysql: &schemasv1alpha4.MysqlSQLTableSchema{},
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
				Postgres: &schemasv1alpha4.SQLTableSchema{},
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
				Mysql: &schemasv1alpha4.MysqlSQLTableSchema{},
			},
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := checkDatabaseTypeMatches(&test.connection, &test.tableSchema)
			assert.Equal(t, test.expect, actual)
		})
	}
}
