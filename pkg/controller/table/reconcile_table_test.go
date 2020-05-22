package table

import (
	"testing"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
)

func Test_checkDatabaseTypeMatches(t *testing.T) {
	tests := []struct {
		name        string
		connection  databasesv1alpha3.DatabaseConnection
		tableSchema schemasv1alpha3.TableSchema
		expect      bool
	}{
		{
			name: "is postgres",
			connection: databasesv1alpha3.DatabaseConnection{
				Postgres: &databasesv1alpha3.PostgresConnection{
					URI: databasesv1alpha3.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha3.TableSchema{
				Postgres: &schemasv1alpha3.SQLTableSchema{},
			},
			expect: true,
		},
		{
			name: "is mysql",
			connection: databasesv1alpha3.DatabaseConnection{
				Mysql: &databasesv1alpha3.MysqlConnection{
					URI: databasesv1alpha3.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha3.TableSchema{
				Mysql: &schemasv1alpha3.SQLTableSchema{},
			},
			expect: true,
		},
		{
			name: "is mysql, expect postgres",
			connection: databasesv1alpha3.DatabaseConnection{
				Mysql: &databasesv1alpha3.MysqlConnection{
					URI: databasesv1alpha3.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha3.TableSchema{
				Postgres: &schemasv1alpha3.SQLTableSchema{},
			},
			expect: false,
		},
		{
			name: "is postgres, expect mysql",
			connection: databasesv1alpha3.DatabaseConnection{
				Postgres: &databasesv1alpha3.PostgresConnection{
					URI: databasesv1alpha3.ValueOrValueFrom{
						Value: "test",
					},
				},
			},
			tableSchema: schemasv1alpha3.TableSchema{
				Mysql: &schemasv1alpha3.SQLTableSchema{},
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
