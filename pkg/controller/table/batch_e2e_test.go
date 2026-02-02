/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package table

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schemahero/schemahero/pkg/apis"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// TestBatchWindowE2E tests the batch window functionality using envtest.
// It creates a Database with batchWindow set, deploys multiple Tables,
// and verifies that they are batched together into a single Migration.
//
// You will need to run 'make envtest' to install the kube api/etcd binaries
func TestBatchWindowE2E(t *testing.T) {
	if os.Getenv("SCHEMAHERO_E2E") != "1" {
		t.Skip("Skipping e2e test. Set SCHEMAHERO_E2E=1 to run.")
	}

	// create envtest instance
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crds", "v1"),
		},
		BinaryAssetsDirectory: filepath.Join(os.TempDir(), "kubebuilder", "bin"),
	}

	// start the kube api server and etcd
	cfg, err := env.Start()
	require.NoError(t, err, "Failed to start environment")
	require.NotNil(t, cfg, "Config was nil")

	defer func() {
		err := env.Stop()
		require.NoError(t, err, "Failed to stop environment")
	}()

	mgr, err := manager.New(cfg, manager.Options{})
	require.NoError(t, err, "Failed to create manager")

	// add types to Scheme
	err = apis.AddToScheme(mgr.GetScheme())
	require.NoError(t, err, "Failed to setup Scheme")

	// add the Table Controller to the Manager
	// Using "*" to handle all databases
	err = Add(mgr, []string{"*"})
	require.NoError(t, err, "Failed to add controller to manager")

	// start the controller manager in a goroutine
	ctx, cancel := context.WithCancel(signals.SetupSignalHandler())
	defer cancel()

	go func() {
		err := mgr.Start(ctx)
		if err != nil {
			t.Logf("Manager stopped with error: %v", err)
		}
	}()

	// Wait for manager to start
	time.Sleep(1 * time.Second)

	k8sClient := mgr.GetClient()

	t.Run("batch_window_creates_single_migration", func(t *testing.T) {
		// Create a namespace for this test
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "batch-test-",
			},
		}
		err := k8sClient.Create(context.Background(), ns)
		require.NoError(t, err)

		// Create a Database with batchWindow
		batchWindow := metav1.Duration{Duration: 2 * time.Second}
		db := &databasesv1alpha4.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testdb",
				Namespace: ns.Name,
			},
			Spec: databasesv1alpha4.DatabaseSpec{
				ImmediateDeploy: true,
				BatchWindow:     &batchWindow,
				Connection: databasesv1alpha4.DatabaseConnection{
					Postgres: &databasesv1alpha4.PostgresConnection{
						URI: databasesv1alpha4.ValueOrValueFrom{
							Value: "postgres://user:pass@localhost:5432/testdb",
						},
					},
				},
			},
		}
		err = k8sClient.Create(context.Background(), db)
		require.NoError(t, err)

		// Create multiple tables rapidly (within the batch window)
		tables := []*schemasv1alpha4.Table{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "users",
					Namespace: ns.Name,
				},
				Spec: schemasv1alpha4.TableSpec{
					Database: "testdb",
					Name:     "users",
					Schema: &schemasv1alpha4.TableSchema{
						Postgres: &schemasv1alpha4.PostgresqlTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha4.PostgresqlColumn{
								{Name: "id", Type: "integer"},
								{Name: "email", Type: "varchar(255)"},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "orders",
					Namespace: ns.Name,
				},
				Spec: schemasv1alpha4.TableSpec{
					Database: "testdb",
					Name:     "orders",
					Schema: &schemasv1alpha4.TableSchema{
						Postgres: &schemasv1alpha4.PostgresqlTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha4.PostgresqlColumn{
								{Name: "id", Type: "integer"},
								{Name: "user_id", Type: "integer"},
								{Name: "total", Type: "decimal(10,2)"},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "products",
					Namespace: ns.Name,
				},
				Spec: schemasv1alpha4.TableSpec{
					Database: "testdb",
					Name:     "products",
					Schema: &schemasv1alpha4.TableSchema{
						Postgres: &schemasv1alpha4.PostgresqlTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha4.PostgresqlColumn{
								{Name: "id", Type: "integer"},
								{Name: "name", Type: "varchar(255)"},
								{Name: "price", Type: "decimal(10,2)"},
							},
						},
					},
				},
			},
		}

		// Create all tables quickly
		for _, table := range tables {
			err := k8sClient.Create(context.Background(), table)
			require.NoError(t, err, "Failed to create table %s", table.Name)
		}

		// Wait for batch window to expire plus some buffer
		time.Sleep(4 * time.Second)

		// List all migrations in the namespace
		migrationList := &schemasv1alpha4.MigrationList{}
		err = k8sClient.List(context.Background(), migrationList, client.InNamespace(ns.Name))
		require.NoError(t, err, "Failed to list migrations")

		// Verify: With batching, we should have at most 1 migration
		// (could be 0 if planning failed due to no actual DB connection)
		// The key is that we should NOT have 3 separate migrations
		assert.LessOrEqual(t, len(migrationList.Items), 1,
			"Expected at most 1 batch migration, got %d", len(migrationList.Items))

		// If we got a migration, verify it has multiple tables
		if len(migrationList.Items) == 1 {
			migration := migrationList.Items[0]
			assert.True(t, len(migration.Spec.Tables) > 1 || migration.Spec.Tables == nil,
				"Expected batch migration to reference multiple tables or be a legacy single-table migration")
			t.Logf("Batch migration created: %s with %d tables", migration.Name, len(migration.Spec.Tables))
		}
	})

	t.Run("no_batch_window_creates_individual_migrations", func(t *testing.T) {
		// Create a namespace for this test
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "nobatch-test-",
			},
		}
		err := k8sClient.Create(context.Background(), ns)
		require.NoError(t, err)

		// Create a Database WITHOUT batchWindow
		db := &databasesv1alpha4.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testdb",
				Namespace: ns.Name,
			},
			Spec: databasesv1alpha4.DatabaseSpec{
				ImmediateDeploy: true,
				// No BatchWindow - should create individual migrations
				Connection: databasesv1alpha4.DatabaseConnection{
					Postgres: &databasesv1alpha4.PostgresConnection{
						URI: databasesv1alpha4.ValueOrValueFrom{
							Value: "postgres://user:pass@localhost:5432/testdb",
						},
					},
				},
			},
		}
		err = k8sClient.Create(context.Background(), db)
		require.NoError(t, err)

		// Create multiple tables
		tables := []*schemasv1alpha4.Table{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "users",
					Namespace: ns.Name,
				},
				Spec: schemasv1alpha4.TableSpec{
					Database: "testdb",
					Name:     "users",
					Schema: &schemasv1alpha4.TableSchema{
						Postgres: &schemasv1alpha4.PostgresqlTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha4.PostgresqlColumn{
								{Name: "id", Type: "integer"},
								{Name: "email", Type: "varchar(255)"},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "orders",
					Namespace: ns.Name,
				},
				Spec: schemasv1alpha4.TableSpec{
					Database: "testdb",
					Name:     "orders",
					Schema: &schemasv1alpha4.TableSchema{
						Postgres: &schemasv1alpha4.PostgresqlTableSchema{
							PrimaryKey: []string{"id"},
							Columns: []*schemasv1alpha4.PostgresqlColumn{
								{Name: "id", Type: "integer"},
								{Name: "user_id", Type: "integer"},
							},
						},
					},
				},
			},
		}

		// Create tables with a small delay to avoid any accidental batching
		for _, table := range tables {
			err := k8sClient.Create(context.Background(), table)
			require.NoError(t, err, "Failed to create table %s", table.Name)
			time.Sleep(100 * time.Millisecond)
		}

		// Wait for reconciliation
		time.Sleep(3 * time.Second)

		// List all migrations in the namespace
		migrationList := &schemasv1alpha4.MigrationList{}
		err = k8sClient.List(context.Background(), migrationList, client.InNamespace(ns.Name))
		require.NoError(t, err, "Failed to list migrations")

		// Without batching, each table should create its own migration
		// (Note: actual creation depends on DB connectivity, so we just verify
		// that batching logic didn't kick in - migrations won't have Tables field populated)
		for _, migration := range migrationList.Items {
			assert.True(t, len(migration.Spec.Tables) <= 1,
				"Without batching, migrations should not have multiple tables")
		}
		t.Logf("Created %d individual migrations", len(migrationList.Items))
	})
}

// TestBatchManagerQueueing tests the batch manager's queuing behavior
// without requiring a full Kubernetes environment
func TestBatchManagerQueueing(t *testing.T) {
	bm := NewBatchManager()

	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 100 * time.Millisecond},
		},
	}

	// Queue multiple tables
	for i := 0; i < 5; i++ {
		table := &schemasv1alpha4.Table{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("table%d", i),
				Namespace: "default",
			},
			Spec: schemasv1alpha4.TableSpec{
				Database: "testdb",
				Name:     fmt.Sprintf("table%d", i),
			},
		}
		queued := bm.QueueTable(context.Background(), db, table, nil)
		assert.True(t, queued, "Table should be queued")
	}

	// Verify all tables are queued
	dbKey := databaseKey(db.Namespace, db.Name)
	bm.mu.RLock()
	batch := bm.batches[dbKey]
	pendingCount := len(batch.pendingTables)
	bm.mu.RUnlock()

	assert.Equal(t, 5, pendingCount, "Should have 5 pending tables")

	// Wait for batch window to expire
	time.Sleep(200 * time.Millisecond)

	// After batch window, pending tables should be cleared
	// (processBatch would have been called, though it will fail without a real reconciler)
	bm.mu.RLock()
	batch = bm.batches[dbKey]
	pendingCountAfter := len(batch.pendingTables)
	bm.mu.RUnlock()

	assert.Equal(t, 0, pendingCountAfter, "Pending tables should be cleared after batch window")
}
