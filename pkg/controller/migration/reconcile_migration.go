package migration

import (
	"context"

	"github.com/pkg/errors"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) reconcileMigration(ctx context.Context, instance *schemasv1alpha3.Migration) (reconcile.Result, error) {
	logger.Debug("reconciling migration",
		zap.String("kind", instance.Kind),
		zap.String("name", instance.Name),
		zap.String("tableName", instance.Spec.TableName))

	if instance.Status.ApprovedAt > 0 && instance.Status.ExecutedAt == 0 {
		table, err := TableFromMigration(ctx, instance)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get table")
		}
		database, err := DatabaseFromTable(ctx, table)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get database")
		}
		connectionURI, err := r.readConnectionURI(database)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get connection uri")
		}

		desiredConfigMap, err := getApplyConfigMap(instance.Name, instance.Namespace, instance.Spec.GeneratedDDL, table.Name, database.Name)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get apply config map")
		}
		var existingConfigMap corev1.ConfigMap
		configMapChanged := false
		err = r.Get(ctx, types.NamespacedName{
			Name:      desiredConfigMap.Name,
			Namespace: desiredConfigMap.Namespace,
		}, &existingConfigMap)
		if kuberneteserrors.IsNotFound(err) {
			// create it
			if err := controllerutil.SetControllerReference(instance, desiredConfigMap, r.scheme); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to set owner on configmap")
			}
			if err := r.Create(ctx, desiredConfigMap); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to create config map")
			}

		} else if err == nil {
			// update it
			existingConfigMap.Data = map[string]string{}
			for k, v := range desiredConfigMap.Data {
				vv, ok := existingConfigMap.Data[k]
				if !ok || vv != v {
					existingConfigMap.Data[k] = v
					configMapChanged = true
				}
			}

			if configMapChanged {
				if err := controllerutil.SetControllerReference(instance, &existingConfigMap, r.scheme); err != nil {
					return reconcile.Result{}, errors.Wrap(err, "failed to update owner on configmap")
				}
				if err = r.Update(ctx, &existingConfigMap); err != nil {
					return reconcile.Result{}, errors.Wrap(err, "failed to update config map")
				}
			}
		} else {
			// something bad is happening here
			return reconcile.Result{}, errors.Wrap(err, "failed to check if config map exists")
		}

		desiredPod, err := getApplyPod(instance.Name, instance.Namespace, connectionURI, database, table)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get apply pod")
		}
		var existingPod corev1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Name:      desiredPod.Name,
			Namespace: desiredPod.Namespace,
		}, &existingPod)
		if kuberneteserrors.IsNotFound(err) {
			// create it
			if err := r.Create(ctx, desiredPod); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to create apply pod")
			}
			if err := controllerutil.SetControllerReference(instance, desiredPod, r.scheme); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to set owner on pod")
			}
		} else if err == nil {
			// maybe update it
			if configMapChanged {
				// restart the pod by deleting and recreating
				logger.Debug("deleting apply pod because config map has changed",
					zap.String("podName", existingPod.Name))
				if err = r.Delete(ctx, &existingPod); err != nil {
					return reconcile.Result{}, errors.Wrap(err, "failed to delete pod")
				}

				// This pod will be recreated later in another exceution of the reconcile loop
				// we watch pods, and when that above delete completed, the reconcile will happen again
			}
		} else {
			// again, something bad
			return reconcile.Result{}, errors.Wrap(err, "failed to check if pod exists")
		}

		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
