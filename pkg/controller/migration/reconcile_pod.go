package migration

import (
	"context"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileMigration) reconcilePod(ctx context.Context, pod *corev1.Pod) (reconcile.Result, error) {
	podLabels := pod.GetObjectMeta().GetLabels()
	role, ok := podLabels["schemahero-role"]
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

	// delete it
	if err := r.Delete(ctx, pod); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to delete apply pod")
	}

	return reconcile.Result{}, nil
}
