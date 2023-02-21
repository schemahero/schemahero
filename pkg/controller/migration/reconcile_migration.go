package migration

import (
	"context"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) reconcileMigration(ctx context.Context, migration *schemasv1alpha4.Migration) (reconcile.Result, error) {
	logger.Debug("checking migration",
		zap.String("name", migration.Name),
		zap.String("tableName", migration.Spec.TableName))

	if !shouldApplyMigration(migration) {
		logger.Debug("migration not yet approved or already executed",
			zap.String("name", migration.Name),
			zap.String("tableName", migration.Spec.TableName))
		return reconcile.Result{}, nil
	}

	databaseInstance, err := getDatabaseFromMigration(ctx, migration)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to get database from migration %s", migration.Name)
	}

	driver, connectionURI, err := databaseInstance.GetConnection(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get connection details for database")
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	statements := db.GetStatementsFromDDL(migration.Spec.GeneratedDDL)

	if err := db.ApplySync(statements); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to apply statements")
	}

	// update the status to applied
	migration.Status.ExecutedAt = time.Now().Unix()
	migration.Status.Phase = schemasv1alpha4.Executed
	err = r.Update(context.Background(), migration)

	if err != nil {
		if kuberneteserrors.IsConflict(err) {
			updatedMigration := &schemasv1alpha4.Migration{}
			err := r.Get(context.Background(), types.NamespacedName{
				Name:      migration.Name,
				Namespace: migration.Namespace,
			}, updatedMigration)
			if err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to get updated instance")
			}

			updatedMigration.Status.ExecutedAt = time.Now().Unix()
			migration.Status.Phase = schemasv1alpha4.Executed
			if err := r.Update(context.Background(), updatedMigration); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to update")
			}
		} else {
			return reconcile.Result{}, errors.Wrap(err, "failed to update")
		}
	}

	return reconcile.Result{}, nil
}

func shouldApplyMigration(migration *schemasv1alpha4.Migration) bool {
	if migration.Status.ApprovedAt > 0 && migration.Status.ExecutedAt == 0 {
		return true
	}
	return false
}

func getDatabaseFromMigration(ctx context.Context, migration *schemasv1alpha4.Migration) (*databasesv1alpha4.Database, error) {
	table, err := TableFromMigration(ctx, migration)
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return nil, errors.Wrap(err, "failed to get table")
		}
	} else {
		database, err := DatabaseFromTable(ctx, table)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get database from table %s", table.Name)
		}
		return database, nil
	}

	view, err := ViewFromMigration(ctx, migration)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get view")
	}
	database, err := DatabaseFromView(ctx, view)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get database from view %s", view.Name)
	}
	return database, nil
}
