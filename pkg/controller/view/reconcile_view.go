package view

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileView is called after filtering events that are not relevant to this
// controller. this function is the main reconcile loop for the view type
func (r *ReconcileView) reconcileView(ctx context.Context, instance *schemasv1alpha4.View) (reconcile.Result, error) {
	logger.Debug("reconciling view",
		zap.String("kind", instance.Kind),
		zap.String("name", instance.Name),
		zap.String("database", instance.Spec.Database),
		zap.String("lastPlannedViewSpecSHA", instance.Status.LastPlannedViewSpecSHA))

	// early exit if the sha of the spec hasn't changed
	currentTableSpecSHA, err := instance.GetSHA()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get instance sha")
	}
	if instance.Status.LastPlannedViewSpecSHA == currentTableSpecSHA {
		return reconcile.Result{}, nil
	}

	// get the full database spec from the api
	database, err := r.getDatabaseInstance(ctx, instance.Namespace, instance.Spec.Database)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database spec")
	}

	// the database object might not yet exist
	// this can happen if the table was deployed at the same time or before the database object
	if database == nil {
		// TDOO add a status field with this state
		logger.Debug("requeuing view reconcile request for 10 seconds because database instance was not present",
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

	// Look for a migration with for this view
	viewSHA, err := instance.GetSHA()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get view sha")
	}

	viewSHA = viewSHA[:7]

	// look for an already calculated migration spec for this view
	migration, err := r.getMigrationSpec(instance.Namespace, viewSHA)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get migration spec")
	}

	if migration != nil {
		// a migration has already been queued. why are we reconciling?
		// if it's not in the sha calculation, then it's not something
		// we are about
		return reconcile.Result{}, nil
	}

	// at this point, we need to execute a plan
	return r.plan(ctx, database, instance)
}

func (r *ReconcileView) getInstance(request reconcile.Request) (*schemasv1alpha4.View, error) {
	v1alpha4instance := &schemasv1alpha4.View{}
	err := r.Get(context.Background(), request.NamespacedName, v1alpha4instance)
	if err != nil {
		return nil, err // don't wrap
	}

	return v1alpha4instance, nil
}

func (r *ReconcileView) getDatabaseInstance(ctx context.Context, namespace string, name string) (*databasesv1alpha4.Database, error) {
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
		// tables might be deployed before a database... if this is the case
		// we don't want to crash, we want to re-reconcile later
		if kuberneteserrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get database object")
	}

	return database, nil
}

func checkDatabaseTypeMatches(connection *databasesv1alpha4.DatabaseConnection, viewSchema *schemasv1alpha4.ViewSchema) bool {
	if connection.Postgres != nil {
		return viewSchema.Postgres != nil
	} else if connection.Mysql != nil {
		return viewSchema.Mysql != nil
	} else if connection.CockroachDB != nil {
		return viewSchema.CockroachDB != nil
	} else if connection.RQLite != nil {
		return viewSchema.RQLite != nil
	} else if connection.TimescaleDB != nil {
		return viewSchema.TimescaleDB != nil
	} else if connection.SQLite != nil {
		return viewSchema.SQLite != nil
	}

	return false
}

// getMigrationSpec will find a migration spec for this exact view object
func (r *ReconcileView) getMigrationSpec(namespace string, viewSHA string) (*schemasv1alpha4.Migration, error) {
	logger.Debug("getting migration spec",
		zap.String("namespace", namespace),
		zap.String("viewSHA", viewSHA))

	// TODO

	return nil, nil
}

// plan will connect to the database and generate a migration spec, deploying the
// migration object
func (r *ReconcileView) plan(ctx context.Context, databaseInstance *databasesv1alpha4.Database, viewInstance *schemasv1alpha4.View) (reconcile.Result, error) {
	logger.Debug("planning migration",
		zap.String("databaseName", databaseInstance.Name),
		zap.String("viewName", viewInstance.Name))

	driver, connectionURI, err := databaseInstance.GetConnection(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get connection details for database")
	}

	db := database.Database{
		Driver:         driver,
		URI:            connectionURI,
		DeploySeedData: databaseInstance.Spec.DeploySeedData,
	}

	schemaStatements, err := db.PlanSyncViewSpec(&viewInstance.Spec)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to plan migration")
	}

	if len(schemaStatements) == 0 {
		logger.Debug("no statements generated for migration",
			zap.String("databaseName", databaseInstance.Name),
			zap.String("viewName", viewInstance.Name))

		return reconcile.Result{}, nil
	}

	viewSHA, err := viewInstance.GetSHA()
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get sha of view")
	}
	viewSHA = viewSHA[:7]

	generatedDDL := strings.Join(schemaStatements, ";\n")

	migration := schemasv1alpha4.Migration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "schemas.schemahero.io/v1alpha4",
			Kind:       "Migration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      viewSHA,
			Namespace: viewInstance.Namespace,
		},
		Spec: schemasv1alpha4.MigrationSpec{
			GeneratedDDL:   generatedDDL,
			DatabaseName:   viewInstance.Spec.Database,
			TableName:      viewInstance.Name,
			TableNamespace: viewInstance.Namespace,
		},
		Status: schemasv1alpha4.MigrationStatus{
			PlannedAt: time.Now().Unix(),
			Phase:     schemasv1alpha4.Planned,
		},
	}

	if databaseInstance.Spec.ImmediateDeploy {
		migration.Status.ApprovedAt = time.Now().Unix()
		migration.Status.Phase = schemasv1alpha4.Planned
	}

	var existingMigration schemasv1alpha4.Migration
	err = r.Get(ctx, types.NamespacedName{
		Name:      migration.Name,
		Namespace: migration.Namespace,
	}, &existingMigration)

	if kuberneteserrors.IsNotFound(err) {
		// create it
		if err := controllerutil.SetControllerReference(viewInstance, &migration, r.scheme); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to set owner on miration")
		}

		if err := r.Create(ctx, &migration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to create migration resource")
		}
	} else if err == nil {
		// update it
		existingMigration.Status = migration.Status
		existingMigration.Spec = migration.Spec
		if err = r.Update(ctx, &existingMigration); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update migration resource")
		}
	} else {
		return reconcile.Result{}, errors.Wrap(err, "failed to get existing migration")
	}

	return reconcile.Result{}, nil
}
