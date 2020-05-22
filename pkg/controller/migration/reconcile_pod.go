package migration

import (
	"context"
	"time"

	"github.com/pkg/errors"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) reconcilePod(ctx context.Context, pod *corev1.Pod) (reconcile.Result, error) {
	podLabels := pod.GetObjectMeta().GetLabels()
	role, ok := podLabels["schemahero-role"]
	if !ok {
		return reconcile.Result{}, nil
	}

	migrationID, ok := podLabels["schemahero-name"]
	if !ok {
		return reconcile.Result{}, nil
	}

	if role != "" && role != "apply" {
		// make sure we are filtering to relevant events
		return reconcile.Result{}, nil
	}

	logger.Debug("reconciling schemahero pod",
		zap.String("kind", pod.Kind),
		zap.String("name", pod.Name),
		zap.String("role", role),
		zap.String("podPhase", string(pod.Status.Phase)))

	if pod.Status.Phase != corev1.PodSucceeded {
		return reconcile.Result{}, nil
	}

	// the migration has completed
	// so lets store the executed at timestamp on the status
	var instance schemasv1alpha3.Migration
	err := r.Get(ctx, types.NamespacedName{
		Name:      migrationID,
		Namespace: pod.Namespace,
	}, &instance)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get migration")
	}
	instance.Status.ExecutedAt = time.Now().Unix()
	if err := r.Update(ctx, &instance); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to update migration")
	}

	// delete related configmaps
	// we don't have owner refs on this because we create them in the wrong order
	// it would be better and more reliable to change this so that the config map
	// has an owner ref set to the pod on creation
	for _, vol := range pod.Spec.Volumes {
		if vol.Name == "input" {
			configMapVolumeSource := vol.VolumeSource.ConfigMap
			if configMapVolumeSource != nil {
				var configMap corev1.ConfigMap
				err := r.Get(ctx, types.NamespacedName{
					Name:      configMapVolumeSource.Name,
					Namespace: pod.Namespace,
				}, &configMap)
				if err == nil {
					if err := r.Delete(ctx, &configMap); err != nil {
						return reconcile.Result{}, errors.Wrap(err, "failed to delete apply config map")
					}
				}
			}
		}
	}

	// delete it
	if err := r.Delete(ctx, pod); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to delete apply pod")
	}

	return reconcile.Result{}, nil
}
