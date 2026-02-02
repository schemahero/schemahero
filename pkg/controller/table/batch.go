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
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// pendingTable represents a table waiting to be processed in a batch
type pendingTable struct {
	table    *schemasv1alpha4.Table
	queuedAt time.Time
}

// batchState tracks the state of a batch for a specific database
type batchState struct {
	pendingTables map[string]*pendingTable // key is namespace/name
	timer         *time.Timer
	mu            sync.Mutex
}

// BatchManager manages batched table processing for databases with BatchWindow configured
type BatchManager struct {
	batches map[string]*batchState // key is database namespace/name
	mu      sync.RWMutex
	client  client.Client
	scheme  *runtime.Scheme
}

// NewBatchManager creates a new BatchManager
func NewBatchManager() *BatchManager {
	return &BatchManager{
		batches: make(map[string]*batchState),
	}
}

// SetClient sets the Kubernetes client for the batch manager
func (bm *BatchManager) SetClient(c client.Client) {
	bm.client = c
}

// SetScheme sets the runtime scheme for the batch manager
func (bm *BatchManager) SetScheme(s *runtime.Scheme) {
	bm.scheme = s
}

// databaseKey returns a unique key for a database
func databaseKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// tableKey returns a unique key for a table
func tableKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// QueueTable adds a table to the batch queue for its database
// Returns true if the table was queued, false if batching is not enabled
func (bm *BatchManager) QueueTable(ctx context.Context, db *databasesv1alpha4.Database, table *schemasv1alpha4.Table, r *ReconcileTable) bool {
	if db.Spec.BatchWindow == nil || db.Spec.BatchWindow.Duration == 0 {
		return false
	}

	dbKey := databaseKey(db.Namespace, db.Name)
	tblKey := tableKey(table.Namespace, table.Name)

	bm.mu.Lock()
	batch, exists := bm.batches[dbKey]
	if !exists {
		batch = &batchState{
			pendingTables: make(map[string]*pendingTable),
		}
		bm.batches[dbKey] = batch
	}
	bm.mu.Unlock()

	batch.mu.Lock()
	defer batch.mu.Unlock()

	// Add or update the pending table
	batch.pendingTables[tblKey] = &pendingTable{
		table:    table.DeepCopy(),
		queuedAt: time.Now(),
	}

	logger.Info("table queued for batch processing",
		zap.String("database", db.Name),
		zap.String("table", table.Name),
		zap.Int("pendingCount", len(batch.pendingTables)),
		zap.Duration("batchWindow", db.Spec.BatchWindow.Duration))

	// Reset or start the timer
	if batch.timer != nil {
		batch.timer.Stop()
	}

	batch.timer = time.AfterFunc(db.Spec.BatchWindow.Duration, func() {
		bm.processBatch(ctx, db, r)
	})

	return true
}

// processBatch processes all pending tables for a database
func (bm *BatchManager) processBatch(ctx context.Context, db *databasesv1alpha4.Database, r *ReconcileTable) {
	dbKey := databaseKey(db.Namespace, db.Name)

	bm.mu.RLock()
	batch, exists := bm.batches[dbKey]
	bm.mu.RUnlock()

	if !exists {
		return
	}

	batch.mu.Lock()
	// Copy and clear pending tables
	tables := make([]*schemasv1alpha4.Table, 0, len(batch.pendingTables))
	for _, pt := range batch.pendingTables {
		tables = append(tables, pt.table)
	}
	batch.pendingTables = make(map[string]*pendingTable)
	batch.timer = nil
	batch.mu.Unlock()

	if len(tables) == 0 {
		return
	}

	logger.Info("processing batch",
		zap.String("database", db.Name),
		zap.Int("tableCount", len(tables)))

	err := r.planBatch(ctx, db, tables)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to process batch"))
	}
}

// planBatch plans multiple tables together and creates a single migration
func (r *ReconcileTable) planBatch(ctx context.Context, databaseInstance *databasesv1alpha4.Database, tables []*schemasv1alpha4.Table) error {
	logger.Info("planning batch migration",
		zap.String("databaseName", databaseInstance.Name),
		zap.Int("tableCount", len(tables)))

	driver, connectionURI, err := databaseInstance.GetConnection(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get connection details for database")
	}

	db := database.Database{
		Driver:         driver,
		URI:            connectionURI,
		DeploySeedData: databaseInstance.Spec.DeploySeedData,
	}

	// Set plugin manager for automatic plugin downloading
	db.SetPluginManager(plugin.GetGlobalPluginManager())

	// Collect all statements and table references
	var allStatements []string
	var tableRefs []schemasv1alpha4.TableReference
	var processedTables []*schemasv1alpha4.Table
	var shaInputs []string

	for _, tableInstance := range tables {
		// plan the schema
		schemaStatements, err := db.PlanSyncTableSpec(&tableInstance.Spec)
		if err != nil {
			logger.Error(errors.Wrapf(err, "failed to plan sync for table %s", tableInstance.Name))
			continue
		}

		// plan the seed data
		seedStatements := []string{}
		if databaseInstance.Spec.DeploySeedData && tableInstance.Spec.SeedData != nil && len(schemaStatements) == 0 {
			stmts, err := db.PlanSyncSeedData(&tableInstance.Spec)
			if err != nil {
				logger.Error(errors.Wrapf(err, "failed to plan seed for table %s", tableInstance.Name))
				continue
			}
			seedStatements = append(seedStatements, stmts...)
		}

		if len(schemaStatements) == 0 && len(seedStatements) == 0 {
			logger.Info("table is already in sync with the desired schema, no migration needed",
				zap.String("databaseName", databaseInstance.Name),
				zap.String("tableName", tableInstance.Name))

			// Update status even if no changes needed
			tableSpecSHA, err := tableInstance.GetSHA()
			if err == nil {
				tableInstance.Status.LastPlannedTableSpecSHA = tableSpecSHA
				if err := r.Status().Update(ctx, tableInstance); err != nil {
					logger.Error(errors.Wrap(err, "failed to update table status"))
				}
			}
			continue
		}

		allStatements = append(allStatements, schemaStatements...)
		allStatements = append(allStatements, seedStatements...)

		tableRefs = append(tableRefs, schemasv1alpha4.TableReference{
			Name:      tableInstance.Name,
			Namespace: tableInstance.Namespace,
		})
		processedTables = append(processedTables, tableInstance)

		// Include table SHA in batch SHA calculation
		tableSHA, _ := tableInstance.GetSHA()
		shaInputs = append(shaInputs, tableSHA)
	}

	if len(allStatements) == 0 {
		logger.Info("no changes needed for any table in batch",
			zap.String("databaseName", databaseInstance.Name))
		return nil
	}

	// Generate a unique SHA for the batch migration
	batchSHAInput := strings.Join(shaInputs, ":")
	batchSHA := fmt.Sprintf("%x", sha256.Sum256([]byte(batchSHAInput)))[:7]

	generatedDDL := strings.Join(allStatements, ";\n")

	migration := schemasv1alpha4.Migration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "schemas.schemahero.io/v1alpha4",
			Kind:       "Migration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("batch-%s", batchSHA),
			Namespace: databaseInstance.Namespace,
		},
		Spec: schemasv1alpha4.MigrationSpec{
			GeneratedDDL:   generatedDDL,
			DatabaseName:   databaseInstance.Name,
			TableName:      processedTables[0].Name, // Primary table for backwards compat
			TableNamespace: processedTables[0].Namespace,
			Tables:         tableRefs,
		},
		Status: schemasv1alpha4.MigrationStatus{
			PlannedAt: time.Now().Unix(),
			Phase:     schemasv1alpha4.Planned,
		},
	}

	if databaseInstance.Spec.ImmediateDeploy {
		migration.Status.ApprovedAt = time.Now().Unix()
		migration.Status.Phase = schemasv1alpha4.Planned
	}

	// Create or update the migration
	var existingMigration schemasv1alpha4.Migration
	err = r.Get(ctx, types.NamespacedName{
		Name:      migration.Name,
		Namespace: migration.Namespace,
	}, &existingMigration)

	if kuberneteserrors.IsNotFound(err) {
		// Set owner reference to first table (for garbage collection)
		if err := controllerutil.SetControllerReference(processedTables[0], &migration, r.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner on migration")
		}

		if err := r.Create(ctx, &migration); err != nil {
			return errors.Wrap(err, "failed to create migration resource")
		}

		logger.Info("created batch migration",
			zap.String("migrationName", migration.Name),
			zap.Int("tableCount", len(tableRefs)))
	} else if err == nil {
		// update it
		existingMigration.Status = migration.Status
		existingMigration.Spec = migration.Spec
		if err = r.Update(ctx, &existingMigration); err != nil {
			return errors.Wrap(err, "failed to update migration resource")
		}
	} else {
		return errors.Wrap(err, "failed to get existing migration")
	}

	// Update status for all processed tables
	for _, tableInstance := range processedTables {
		tableSpecSHA, err := tableInstance.GetSHA()
		if err != nil {
			logger.Error(errors.Wrap(err, "failed to get table sha for status update"))
			continue
		}
		tableInstance.Status.LastPlannedTableSpecSHA = tableSpecSHA
		if err := r.Status().Update(ctx, tableInstance); err != nil {
			logger.Error(errors.Wrap(err, "failed to update table status"))
		}
	}

	return nil
}

// global batch manager instance
var globalBatchManager = NewBatchManager()

// GetBatchManager returns the global batch manager
func GetBatchManager() *BatchManager {
	return globalBatchManager
}
