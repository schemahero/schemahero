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
	"testing"
	"time"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_databaseKey(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		dbName    string
		expected  string
	}{
		{
			name:      "basic key",
			namespace: "default",
			dbName:    "mydb",
			expected:  "default/mydb",
		},
		{
			name:      "different namespace",
			namespace: "production",
			dbName:    "app-db",
			expected:  "production/app-db",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := databaseKey(test.namespace, test.dbName)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_tableKey(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		tableName string
		expected  string
	}{
		{
			name:      "basic key",
			namespace: "default",
			tableName: "users",
			expected:  "default/users",
		},
		{
			name:      "different namespace",
			namespace: "myapp",
			tableName: "orders",
			expected:  "myapp/orders",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := tableKey(test.namespace, test.tableName)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_BatchManager_NoBatchingWithoutBatchWindow(t *testing.T) {
	bm := NewBatchManager()

	// Database without BatchWindow
	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			// No BatchWindow set
		},
	}

	table := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "users",
			Namespace: "default",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "testdb",
			Name:     "users",
		},
	}

	// Should return false (not queued) when BatchWindow is not set
	queued := bm.QueueTable(nil, db, table, nil)
	assert.False(t, queued, "table should not be queued when BatchWindow is not set")
}

func Test_BatchManager_NoBatchingWithZeroBatchWindow(t *testing.T) {
	bm := NewBatchManager()

	// Database with zero BatchWindow
	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 0},
		},
	}

	table := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "users",
			Namespace: "default",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "testdb",
			Name:     "users",
		},
	}

	// Should return false (not queued) when BatchWindow is zero
	queued := bm.QueueTable(nil, db, table, nil)
	assert.False(t, queued, "table should not be queued when BatchWindow is zero")
}

func Test_BatchManager_QueuesTableWithBatchWindow(t *testing.T) {
	bm := NewBatchManager()

	// Database with BatchWindow
	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 5 * time.Second},
		},
	}

	table := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "users",
			Namespace: "default",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "testdb",
			Name:     "users",
		},
	}

	// Should return true (queued) when BatchWindow is set
	queued := bm.QueueTable(nil, db, table, nil)
	assert.True(t, queued, "table should be queued when BatchWindow is set")

	// Verify the table is in the pending list
	dbKey := databaseKey(db.Namespace, db.Name)
	bm.mu.RLock()
	batch, exists := bm.batches[dbKey]
	bm.mu.RUnlock()

	assert.True(t, exists, "batch should exist for database")
	assert.Equal(t, 1, len(batch.pendingTables), "should have 1 pending table")
}

func Test_BatchManager_QueuesMultipleTables(t *testing.T) {
	bm := NewBatchManager()

	// Database with BatchWindow
	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 5 * time.Second},
		},
	}

	tables := []*schemasv1alpha4.Table{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "users",
				Namespace: "default",
			},
			Spec: schemasv1alpha4.TableSpec{
				Database: "testdb",
				Name:     "users",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "orders",
				Namespace: "default",
			},
			Spec: schemasv1alpha4.TableSpec{
				Database: "testdb",
				Name:     "orders",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "products",
				Namespace: "default",
			},
			Spec: schemasv1alpha4.TableSpec{
				Database: "testdb",
				Name:     "products",
			},
		},
	}

	// Queue all tables
	for _, table := range tables {
		queued := bm.QueueTable(nil, db, table, nil)
		assert.True(t, queued, "table should be queued")
	}

	// Verify all tables are in the pending list
	dbKey := databaseKey(db.Namespace, db.Name)
	bm.mu.RLock()
	batch, exists := bm.batches[dbKey]
	bm.mu.RUnlock()

	assert.True(t, exists, "batch should exist for database")
	assert.Equal(t, 3, len(batch.pendingTables), "should have 3 pending tables")
}

func Test_BatchManager_SeparateBatchesPerDatabase(t *testing.T) {
	bm := NewBatchManager()

	// Two different databases
	db1 := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db1",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 5 * time.Second},
		},
	}

	db2 := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db2",
			Namespace: "default",
		},
		Spec: databasesv1alpha4.DatabaseSpec{
			BatchWindow: &metav1.Duration{Duration: 5 * time.Second},
		},
	}

	table1 := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "users",
			Namespace: "default",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "db1",
			Name:     "users",
		},
	}

	table2 := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orders",
			Namespace: "default",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "db2",
			Name:     "orders",
		},
	}

	bm.QueueTable(nil, db1, table1, nil)
	bm.QueueTable(nil, db2, table2, nil)

	// Verify separate batches exist
	bm.mu.RLock()
	batch1, exists1 := bm.batches[databaseKey(db1.Namespace, db1.Name)]
	batch2, exists2 := bm.batches[databaseKey(db2.Namespace, db2.Name)]
	bm.mu.RUnlock()

	assert.True(t, exists1, "batch should exist for db1")
	assert.True(t, exists2, "batch should exist for db2")
	assert.Equal(t, 1, len(batch1.pendingTables), "db1 should have 1 pending table")
	assert.Equal(t, 1, len(batch2.pendingTables), "db2 should have 1 pending table")
}
