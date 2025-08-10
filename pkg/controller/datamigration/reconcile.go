/*
Copyright 2025 The SchemaHero Authors

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

package datamigration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileDataMigration) reconcileDataMigration(ctx context.Context, instance *schemasv1alpha4.DataMigration) (reconcile.Result, error) {
	logger.Debug("reconciling datamigration",
		zap.String("name", instance.Name),
		zap.String("database", instance.Spec.Database))

	// Get the database instance
	database, err := r.getDatabaseInstance(ctx, instance.Namespace, instance.Spec.Database)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database spec")
	}

	// The database object might not yet exist
	if database == nil {
		logger.Debug("requeuing datamigration reconcile request because database instance was not present",
			zap.String("database.name", instance.Spec.Database),
			zap.String("database.namespace", instance.Namespace))

		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}, nil
	}

	// Generate SQL for the data migrations
	statements, err := r.generateSQL(database, instance)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to generate SQL")
	}

	if len(statements) == 0 {
		logger.Debug("no statements generated for data migration",
			zap.String("name", instance.Name))
		return reconcile.Result{}, nil
	}

	// Create a Migration resource
	return r.createMigration(ctx, database, instance, statements)
}

func (r *ReconcileDataMigration) getDatabaseInstance(ctx context.Context, namespace string, name string) (*databasesv1alpha4.Database, error) {
	logger.Debug("getting database spec",
		zap.String("namespace", namespace),
		zap.String("name", name))

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	
	databasesClient, err := databasesclientv1alpha4.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databasesclient")
	}

	database, err := databasesClient.Databases(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get database object")
	}

	return database, nil
}

func (r *ReconcileDataMigration) generateSQL(database *databasesv1alpha4.Database, dataMigration *schemasv1alpha4.DataMigration) ([]string, error) {
	statements := []string{}

	// Determine database type
	var dbType string
	if database.Spec.Connection.Postgres != nil || database.Spec.Connection.CockroachDB != nil || database.Spec.Connection.TimescaleDB != nil {
		dbType = "postgres"
	} else if database.Spec.Connection.Mysql != nil {
		dbType = "mysql"
	} else {
		return nil, fmt.Errorf("unsupported database type for data migrations")
	}

	// Generate SQL for each migration operation
	for _, migration := range dataMigration.Spec.Migrations {
		sql, err := generateMigrationSQL(dbType, migration)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate SQL for migration type %s", migration.Type)
		}
		if sql != "" {
			statements = append(statements, sql)
		}
	}

	return statements, nil
}

func generateMigrationSQL(dbType string, migration schemasv1alpha4.DataMigrationOperation) (string, error) {
	switch migration.Type {
	case schemasv1alpha4.UpdateMigration:
		return generateUpdateSQL(dbType, migration)
	case schemasv1alpha4.CalculateMigration:
		return generateCalculateSQL(dbType, migration)
	case schemasv1alpha4.ConvertMigration:
		return generateConvertSQL(dbType, migration)
	default:
		return "", fmt.Errorf("unknown migration type: %s", migration.Type)
	}
}

func generateUpdateSQL(dbType string, migration schemasv1alpha4.DataMigrationOperation) (string, error) {
	if migration.Table == "" || migration.Column == "" || migration.Value == "" {
		return "", fmt.Errorf("update migration requires table, column, and value")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s = %s", migration.Table, migration.Column, migration.Value)
	if migration.Where != "" {
		sql += fmt.Sprintf(" WHERE %s", migration.Where)
	}

	return sql, nil
}

func generateCalculateSQL(dbType string, migration schemasv1alpha4.DataMigrationOperation) (string, error) {
	if migration.Table == "" || migration.Column == "" || migration.Expression == "" {
		return "", fmt.Errorf("calculate migration requires table, column, and expression")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s = %s", migration.Table, migration.Column, migration.Expression)
	if migration.Where != "" {
		sql += fmt.Sprintf(" WHERE %s", migration.Where)
	}

	return sql, nil
}

func generateConvertSQL(dbType string, migration schemasv1alpha4.DataMigrationOperation) (string, error) {
	if migration.Table == "" || migration.Column == "" || migration.From == "" || migration.To == "" {
		return "", fmt.Errorf("convert migration requires table, column, from, and to")
	}

	// Handle common type conversions
	if dbType == "postgres" {
		if migration.From == "timestamp" && migration.To == "timestamptz" {
			return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE timestamptz USING %s AT TIME ZONE 'UTC'",
				migration.Table, migration.Column, migration.Column), nil
		}
	} else if dbType == "mysql" {
		// MySQL has different syntax for type conversions
		if migration.From == "timestamp" && migration.To == "timestamptz" {
			// MySQL doesn't have timestamptz, but we can convert to DATETIME
			return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s DATETIME",
				migration.Table, migration.Column), nil
		}
	}

	// Generic conversion
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
		migration.Table, migration.Column, migration.To), nil
}

func (r *ReconcileDataMigration) createMigration(ctx context.Context, database *databasesv1alpha4.Database, dataMigration *schemasv1alpha4.DataMigration, statements []string) (reconcile.Result, error) {
	// Generate a unique name for the migration
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s", dataMigration.Namespace, dataMigration.Name)))
	migrationName := fmt.Sprintf("datamig-%x", hash)[:16]

	generatedDDL := strings.Join(statements, ";\n")

	migration := schemasv1alpha4.Migration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "schemas.schemahero.io/v1alpha4",
			Kind:       "Migration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      migrationName,
			Namespace: dataMigration.Namespace,
		},
		Spec: schemasv1alpha4.MigrationSpec{
			GeneratedDDL:   generatedDDL,
			DatabaseName:   dataMigration.Spec.Database,
			TableName:      "data-migration",
			TableNamespace: dataMigration.Namespace,
		},
		Status: schemasv1alpha4.MigrationStatus{
			PlannedAt: time.Now().Unix(),
			Phase:     schemasv1alpha4.Planned,
		},
	}

	// If immediate deploy is enabled, auto-approve
	if database.Spec.ImmediateDeploy {
		migration.Status.ApprovedAt = time.Now().Unix()
		migration.Status.Phase = schemasv1alpha4.Approved
	}

	// Check if migration already exists
	var existingMigration schemasv1alpha4.Migration
	err := r.Get(ctx, types.NamespacedName{
		Name:      migration.Name,
		Namespace: migration.Namespace,
	}, &existingMigration)

	if kuberneteserrors.IsNotFound(err) {
		// Create the migration
		if err := controllerutil.SetControllerReference(dataMigration, &migration, r.scheme); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to set owner on migration")
		}

		if err := r.Create(ctx, &migration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to create migration resource")
		}

		// Update DataMigration status
		dataMigration.Status.MigrationName = migrationName
		dataMigration.Status.Phase = schemasv1alpha4.Planned
		dataMigration.Status.PlannedAt = time.Now().Unix()
		
		if err := r.Status().Update(ctx, dataMigration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update datamigration status")
		}
	} else if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get existing migration")
	}

	logger.Info("created migration for data migration",
		zap.String("datamigration", dataMigration.Name),
		zap.String("migration", migrationName))

	return reconcile.Result{}, nil
}