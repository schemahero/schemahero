package table

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	databasesclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha3"
	schemasclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileTable) reconcilePod(ctx context.Context, pod *corev1.Pod) (reconcile.Result, error) {
	podLabels := pod.GetObjectMeta().GetLabels()
	role, ok := podLabels["schemahero-role"]
	if !ok {
		return reconcile.Result{}, nil
	}

	if role != "table" && role != "plan" {
		// we want to avoid migration pods in this reconciler
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

	// Write the plan from stdout to the object itself
	cfg, err := config.GetConfig()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get config")
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to create client")
	}

	podLogOpts := corev1.PodLogOptions{}
	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to open log stream")
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to copy logs too buffer")
	}

	out := buf.String()

	// remove empty lines from output
	// the planner plans each row, and can leave empty lines
	out = strings.Replace(out, "\n\n", "\n", -1)

	logger.Debug("read output from pod",
		zap.String("kind", pod.Kind),
		zap.String("name", pod.Name),
		zap.String("role", role),
		zap.String("output", out))

	tableName, ok := podLabels["schemahero-name"]
	if !ok {
		return reconcile.Result{}, nil
	}
	tableNamespace, ok := podLabels["schemahero-namespace"]
	if !ok {
		return reconcile.Result{}, nil
	}

	schemasClient, err := schemasclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to create schema client")
	}

	table, err := schemasClient.Tables(tableNamespace).Get(ctx, tableName, metav1.GetOptions{})
	if err != nil {
		// this isn't an error condition, the table could have been deleted before
		// the table pod reconciled
		// there's no reason to log something here, the correct behavior is to do nothing
		logger.Info("table not found while reconiling pod",
			zap.String("namespace", tableNamespace),
			zap.String("name", tableName))
		return reconcile.Result{}, nil
	}
	tableSHA, err := table.GetSHA()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get sha of table")
	}

	desiredMigration := schemasv1alpha3.Migration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "schemas.schemahero.io/v1alpha3",
			Kind:       "Migration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tableSHA,
			Namespace: table.Namespace,
		},
		Spec: schemasv1alpha3.MigrationSpec{
			GeneratedDDL:   out,
			TableName:      table.Name,
			TableNamespace: table.Namespace,
		},
		Status: schemasv1alpha3.MigrationStatus{
			PlannedAt: time.Now().Unix(),
		},
	}

	// If the database is set to immediate deploy, then set it as approved also
	databasesClient, err := databasesclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Failed to create database client")
	}

	database, err := databasesClient.Databases(tableNamespace).Get(ctx, table.Spec.Database, metav1.GetOptions{})
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database")
	}

	if database.Spec.ImmediateDeploy {
		desiredMigration.Status.ApprovedAt = time.Now().Unix()
	}

	var existingMigration schemasv1alpha3.Migration
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredMigration.Name,
		Namespace: desiredMigration.Namespace,
	}, &existingMigration)
	if kuberneteserrors.IsNotFound(err) {
		// create it
		if err := controllerutil.SetControllerReference(table, &desiredMigration, r.scheme); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to set owner on miration")
		}

		if err := r.Create(ctx, &desiredMigration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to create migration resource")
		}
	} else if err == nil {
		// update it
		existingMigration.Status = desiredMigration.Status
		existingMigration.Spec = desiredMigration.Spec
		if err = r.Update(ctx, &existingMigration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update migration resource")
		}
	} else {
		return reconcile.Result{}, errors.Wrap(err, "failed to get existing migration")
	}

	// Delete the pod and config map
	if err := r.Delete(ctx, pod); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to delete pod from plan phase")
	}

	configMapName := ""
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == "specs" && volume.ConfigMap != nil {
			configMapName = volume.ConfigMap.Name
		}
	}
	configMap := corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: pod.Namespace}, &configMap)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get config map from plan phase")
	}

	if err := r.Delete(ctx, &configMap); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to delete config map from plan phase")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTable) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha3.ValueOrValueFrom) (string, error) {
	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	if valueOrValueFrom.ValueFrom == nil {
		return "", errors.New("value and valueFrom cannot both be nil/empty")
	}

	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Name:      valueOrValueFrom.ValueFrom.SecretKeyRef.Name,
			Namespace: namespace,
		}

		if err := r.Get(context.Background(), secretNamespacedName, secret); err != nil {
			if kuberneteserrors.IsNotFound(err) {
				return "", errors.New("table secret not found")
			} else {
				return "", errors.Wrap(err, "failed to get existing connection secret")
			}
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	if valueOrValueFrom.ValueFrom.Vault != nil {
		// this feels wrong, but also doesn't make sense to return a
		// a URI ref as a connection URI?
		return "", nil
	}

	return "", errors.New("unable to find supported valueFrom")
}
