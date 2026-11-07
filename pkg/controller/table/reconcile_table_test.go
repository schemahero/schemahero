package table

import (
	"context"
	"testing"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func Test_reconcileTable_replansWhenMigrationDeleted(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, schemasv1alpha4.AddToScheme(scheme))

	table := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{Name: "users", Namespace: "default"},
		Spec: schemasv1alpha4.TableSpec{
			Database: "db",
			Name:     "users",
		},
	}
	sha, err := table.GetSHA()
	require.NoError(t, err)
	table.Status.LastPlannedTableSpecSHA = sha

	migration := &schemasv1alpha4.Migration{
		ObjectMeta: metav1.ObjectMeta{Name: sha[:7], Namespace: "default"},
	}

	t.Run("migration still exists, skip planning", func(t *testing.T) {
		r := &ReconcileTable{
			Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(migration.DeepCopy()).Build(),
			scheme: scheme,
		}
		exists, err := r.migrationExists(context.Background(), "default", sha[:7])
		require.NoError(t, err)
		assert.True(t, exists)

		_, err = r.reconcileTable(context.Background(), table.DeepCopy())
		assert.NoError(t, err)
	})

	t.Run("migration deleted, plan again", func(t *testing.T) {
		r := &ReconcileTable{
			Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
			scheme: scheme,
		}
		exists, err := r.migrationExists(context.Background(), "default", sha[:7])
		require.NoError(t, err)
		assert.False(t, exists)

		// planning is attempted again, which fails here because there is no
		// cluster to look up the database object in. it must not exit early.
		_, err = r.reconcileTable(context.Background(), table.DeepCopy())
		assert.Error(t, err)
	})
}
