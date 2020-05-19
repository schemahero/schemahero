package table

import (
	"context"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	databasesclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileTable) reconcileTable(ctx context.Context, instance *schemasv1alpha3.Table) (reconcile.Result, error) {
	logger.Debug("reconciling table",
		zap.String("kind", instance.Kind),
		zap.String("name", instance.Name),
		zap.String("database", instance.Spec.Database))

	// get the full database spec from the api
	database, err := r.getDatabaseSpec(ctx, instance.Namespace, instance.Spec.Database)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database spec")
	}

	// the database object might not yet exist
	// this can happen if the table was deployed at the same time or before the database object
	if database == nil {
		// TDOO add a status field with this state
		logger.Debug("requeuing table reconcile request for 10 seconds because database instance was not present",
			zap.String("database.name", instance.Spec.Database),
			zap.String("database.namespace", instance.Namespace))

		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}, nil
	}

	matchingType := checkDatabaseTypeMatches(&database.Spec.Connection, instance.Spec.Schema)
	if !matchingType {
		// TODO add a status field with this state
		return reconcile.Result{}, errors.New("unable to deploy table to connection of different type")
	}

	// Look for a migration with for this table
	tableSHA, err := instance.GetSHA()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get table sha")
	}

	migration, err := r.getMigrationSpec(instance.Namespace, tableSHA)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get migration spec")
	}

	if migration != nil {
		// a migration has already been queued. why are we reconciling?
		// if it's not in the sha calculation, then it's not something
		// we are about
		return reconcile.Result{}, nil
	}

	// Deploy a pod to calculculate the plan
	return r.deployMigrationPlanPhase(ctx, database, instance)
}

func (r *ReconcileTable) getInstance(request reconcile.Request) (*schemasv1alpha3.Table, error) {
	v1alpha3instance := &schemasv1alpha3.Table{}
	err := r.Get(context.Background(), request.NamespacedName, v1alpha3instance)
	if err != nil {
		return nil, err // don't wrap
	}

	return v1alpha3instance, nil
}

func (r *ReconcileTable) getMigrationSpec(namespace string, name string) (*schemasv1alpha3.Migration, error) {
	logger.Debug("getting migration spec",
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil, nil
}

func (r *ReconcileTable) getDatabaseSpec(ctx context.Context, namespace string, name string) (*databasesv1alpha3.Database, error) {
	logger.Debug("getting database spec",
		zap.String("namespace", namespace),
		zap.String("name", name))

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	databasesClient, err := databasesclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	database, err := databasesClient.Databases(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// tables might be deployed before a database... if this is the case
		// we don't want to crash, we want to re-reconcile later
		if kuberneteserrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// try to parse the secret too, the database may be deployed, but that doesn't mean it's ready
	// TODO this would be better as a status field on the database object, instead of this leaky
	// interface
	if database.Spec.Connection.Postgres != nil {
		_, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.Postgres.URI)
		if err != nil {
			return nil, nil
		}
	} else if database.Spec.Connection.Mysql != nil {
		_, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.Mysql.URI)
		if err != nil {
			return nil, nil
		}
	} else if database.Spec.Connection.CockroachDB != nil {
		_, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.CockroachDB.URI)
		if err != nil {
			return nil, nil
		}
	}

	return database, nil
}

func checkDatabaseTypeMatches(connection *databasesv1alpha3.DatabaseConnection, tableSchema *schemasv1alpha3.TableSchema) bool {
	if connection.Postgres != nil {
		return tableSchema.Postgres != nil
	} else if connection.Mysql != nil {
		return tableSchema.Mysql != nil
	} else if connection.CockroachDB != nil {
		return tableSchema.CockroachDB != nil
	}

	return false
}

func (r *ReconcileTable) deployMigrationPlanPhase(ctx context.Context, database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (reconcile.Result, error) {
	logger.Debug("deploying plan phase of migration",
		zap.String("databaseName", database.Name),
		zap.String("tableName", table.Name))

	desiredConfigMap, err := getPlanConfigMap(database, table)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get config map object for plan")
	}
	var existingConfigMap corev1.ConfigMap
	configMapChanged := false
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredConfigMap.Name,
		Namespace: desiredConfigMap.Namespace,
	}, &existingConfigMap)
	if kuberneteserrors.IsNotFound(err) {
		// create it
		if err := controllerutil.SetControllerReference(table, desiredConfigMap, r.scheme); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to set owner on configmap")
		}
		if err := r.Create(ctx, desiredConfigMap); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to create config map")
		}
	} else if err == nil {
		// update it
		existingConfigMap.Data = map[string]string{}
		for k, v := range desiredConfigMap.Data {
			d, ok := existingConfigMap.Data[k]
			if !ok || d != v {
				existingConfigMap.Data[k] = v
				configMapChanged = true
			}
		}
		if configMapChanged {
			if err := controllerutil.SetControllerReference(table, &existingConfigMap, r.scheme); err != nil {
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

	if database.UsingVault() {
		desiredServiceAccount := getPlanServiceAccount(database)

		var sa corev1.ServiceAccount
		err = r.Get(ctx, types.NamespacedName{
			Name:      desiredServiceAccount.Name,
			Namespace: desiredServiceAccount.Namespace,
		}, &sa)
		if kuberneteserrors.IsNotFound(err) {
			// create it
			if err := r.Create(ctx, desiredServiceAccount); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to create plan service account")
			}
			if err := controllerutil.SetControllerReference(table, desiredServiceAccount, r.scheme); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to set owner on service account")
			}
		} else if err != nil {
			// again, something bad
			return reconcile.Result{}, errors.Wrap(err, "failed to check if service account exists")
		}
	}

	desiredPod, err := r.getPlanPod(database, table)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get pod for plan")
	}
	var existingPod corev1.Pod
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredPod.Name,
		Namespace: desiredPod.Namespace,
	}, &existingPod)
	if kuberneteserrors.IsNotFound(err) {
		// create it
		if err := r.Create(ctx, desiredPod); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to create plan pod")
		}
		if err := controllerutil.SetControllerReference(table, desiredPod, r.scheme); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to set owner on pod")
		}
	} else if err == nil {
		// maybe update it
		if configMapChanged {
			// restart the pod by deleting and recreating
			logger.Debug("deleting plan pod because config map has changed",
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
