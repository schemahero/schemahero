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
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileTable) reconcileInstance(instance *schemasv1alpha3.Table) (reconcile.Result, error) {
	logger.Debug("reconciling table",
		zap.String("kind", instance.Kind),
		zap.String("name", instance.Name),
		zap.String("database", instance.Spec.Database))

	database, err := r.getDatabaseSpec(instance.Namespace, instance.Spec.Database)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database spec")
	}

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
	if err := r.plan(database, instance); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to schedule plan phase")
	}

	return reconcile.Result{}, nil
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

func (r *ReconcileTable) getDatabaseSpec(namespace string, name string) (*databasesv1alpha3.Database, error) {
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

	database, err := databasesClient.Databases(namespace).Get(name, metav1.GetOptions{})
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
	}

	return database, nil
}

func checkDatabaseTypeMatches(connection *databasesv1alpha3.DatabaseConnection, tableSchema *schemasv1alpha3.TableSchema) bool {
	if connection.Postgres != nil {
		return tableSchema.Postgres != nil
	} else if connection.Mysql != nil {
		return tableSchema.Mysql != nil
	}

	return false
}
