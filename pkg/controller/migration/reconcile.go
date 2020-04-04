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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) getInstance(request reconcile.Request) (*schemasv1alpha3.Migration, error) {
	v1alpha3instance := &schemasv1alpha3.Migration{}
	err := r.Get(context.Background(), request.NamespacedName, v1alpha3instance)
	if err != nil {
		return nil, err // don't wrap
	}

	return v1alpha3instance, nil
}

func (r *ReconcileMigration) reconcileInstance(ctx context.Context, instance *schemasv1alpha3.Migration) (reconcile.Result, error) {
	logger.Debug("reconciling migration",
		zap.String("kind", instance.Kind),
		zap.String("name", instance.Name),
		zap.String("tableName", instance.Spec.TableName))

	if instance.Status.ApprovedAt > 0 && instance.Status.ExecutedAt == 0 {
		table, err := tableFromMigration(ctx, instance)
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to get table")
		}
		database, err := databaseFromTable(ctx, table)
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
			if err := r.Create(ctx, desiredConfigMap); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to create config map")
			}
		} else if err == nil {
			// update it
			existingConfigMap.Data = map[string]string{}
			for k, v := range desiredConfigMap.Data {
				existingConfigMap.Data[k] = v
			}

			if err = r.Update(ctx, &existingConfigMap); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to update config map")
			}
			configMapChanged = true
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

func (r *ReconcileMigration) reconcilePod(pod *corev1.Pod) (reconcile.Result, error) {
	// podLabels := pod.GetObjectMeta().GetLabels()
	// role, ok := podLabels["schemahero-role"]
	// if !ok {
	// 	return reconcile.Result{}, nil
	// }

	return reconcile.Result{}, nil
}
