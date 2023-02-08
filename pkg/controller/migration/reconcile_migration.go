package migration

import (
	"context"
	"time"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) reconcileMigration(ctx context.Context, instance *schemasv1alpha4.Migration) (reconcile.Result, error) {
	logger.Debug("reconciling migration",
		zap.String("name", instance.Name),
		zap.String("tableName", instance.Spec.TableName))

	if instance.Status.ApprovedAt > 0 && instance.Status.ExecutedAt == 0 {
		tableInstance, err := TableFromMigration(ctx, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}

		databaseInstance, err := DatabaseFromTable(ctx, tableInstance)
		if err != nil {
			return reconcile.Result{}, nil
		}

		driver, connectionURI, err := databaseInstance.GetConnection(ctx)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get connection details for database")
		}

		db := database.Database{
			Driver: driver,
			URI:    connectionURI,
		}

		statements := db.GetStatementsFromDDL(instance.Spec.GeneratedDDL)

		if err := db.ApplySync(statements); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to apply statements")
		}

		// update the status to applied
		instance.Status.ExecutedAt = time.Now().Unix()
		instance.Status.Phase = schemasv1alpha4.Executed
		err = r.Update(context.Background(), instance)

		if err != nil {
			if kuberneteserrors.IsConflict(err) {
				updatedInstance := &schemasv1alpha4.Migration{}
				err := r.Get(context.Background(), types.NamespacedName{
					Name:      instance.Name,
					Namespace: instance.Namespace,
				}, updatedInstance)
				if err != nil {
					return reconcile.Result{}, errors.Wrap(err, "failed to get updated instance")
				}

				updatedInstance.Status.ExecutedAt = time.Now().Unix()
				instance.Status.Phase = schemasv1alpha4.Executed
				if err := r.Update(context.Background(), updatedInstance); err != nil {
					return reconcile.Result{}, errors.Wrap(err, "failed to update")
				}
			} else {
				return reconcile.Result{}, errors.Wrap(err, "failed to update")
			}
		}

	}

	return reconcile.Result{}, nil
}
