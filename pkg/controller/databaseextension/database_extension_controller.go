/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package databaseextension

import (
	"context"
	"slices"
	"time"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDatabaseExtension{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("databaseextension-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &schemasv1alpha4.DatabaseExtension{}, &handler.TypedEnqueueRequestForObject[*schemasv1alpha4.DatabaseExtension]{}))
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDatabaseExtension{}

type ReconcileDatabaseExtension struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileDatabaseExtension) dropExtension(ctx context.Context, databaseExtension *schemasv1alpha4.DatabaseExtension) error {
	logger.Debug("dropping database extension",
		zap.String("name", databaseExtension.Name),
		zap.String("namespace", databaseExtension.Namespace))

	dbInstance, err := r.getDatabaseFromExtension(ctx, databaseExtension)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	driver, connectionURI, err := dbInstance.GetConnection(ctx)
	if err != nil {
		return err
	}

	if driver != "postgres" || databaseExtension.Spec.Postgres == nil {
		return nil
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	// Set plugin manager for automatic plugin downloading
	db.SetPluginManager(plugin.GetGlobalPluginManager())

	// Get a connection to plan the extension drop
	conn, err := db.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Use PlanExtensionSchema with a special marker to indicate drop
	extensionSchema := &schemasv1alpha4.PostgresDatabaseExtension{
		Name:      databaseExtension.Spec.Postgres.Name,
		Version:   databaseExtension.Spec.Postgres.Version,
		IsDeleted: true, // Special marker for drop operation
	}

	statements, err := conn.PlanExtensionSchema(databaseExtension.Spec.Postgres.Name, extensionSchema)
	if err != nil {
		return err
	}

	return db.ApplySync(statements)
}

func (r *ReconcileDatabaseExtension) getDatabaseFromExtension(ctx context.Context, databaseExtension *schemasv1alpha4.DatabaseExtension) (*databasesv1alpha4.Database, error) {
	database := &databasesv1alpha4.Database{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      databaseExtension.Spec.Database,
		Namespace: databaseExtension.Namespace,
	}, database)
	if err != nil {
		return nil, err
	}

	return database, nil
}

// Reconcile reads that state of the cluster for a DatabaseExtension object and makes changes based on the state read
// and what is in the DatabaseExtension.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=databaseextensions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=databaseextensions/status,verbs=get;update;patch
func (r *ReconcileDatabaseExtension) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger.Debug("reconciling database extension",
		zap.String("kind", "databaseextension"),
		zap.String("name", request.Name),
		zap.String("namespace", request.Namespace))

	databaseExtension := &schemasv1alpha4.DatabaseExtension{}
	err := r.Get(ctx, request.NamespacedName, databaseExtension)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	finalizerName := "databaseextensions.schemas.schemahero.io/finalizer"

	if !databaseExtension.ObjectMeta.DeletionTimestamp.IsZero() {
		if databaseExtension.Spec.RemoveOnDeletion && slices.Contains(databaseExtension.ObjectMeta.Finalizers, finalizerName) {
			if err := r.dropExtension(ctx, databaseExtension); err != nil {
				return reconcile.Result{}, err
			}

			databaseExtension.ObjectMeta.Finalizers = removeString(databaseExtension.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, databaseExtension); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	if databaseExtension.Spec.RemoveOnDeletion && !slices.Contains(databaseExtension.ObjectMeta.Finalizers, finalizerName) {
		databaseExtension.ObjectMeta.Finalizers = append(databaseExtension.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, databaseExtension); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	dbInstance, err := r.getDatabaseFromExtension(ctx, databaseExtension)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			logger.Debug("database not found, requeuing", zap.String("database", databaseExtension.Spec.Database))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 10,
			}, nil
		}
		return reconcile.Result{}, err
	}

	driver, connectionURI, err := dbInstance.GetConnection(ctx)
	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}

	if driver != "postgres" || databaseExtension.Spec.Postgres == nil {
		logger.Debug("not a postgres database or no postgres extension specified, skipping")
		return reconcile.Result{}, nil
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	// Set plugin manager for automatic plugin downloading
	db.SetPluginManager(plugin.GetGlobalPluginManager())

	// Get a connection to plan the extension creation
	conn, err := db.GetConnection(ctx)
	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}
	defer conn.Close()

	// Use PlanExtensionSchema to create statements
	statements, err := conn.PlanExtensionSchema(databaseExtension.Spec.Postgres.Name, databaseExtension.Spec.Postgres)
	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}

	if err := db.ApplySync(statements); err != nil {
		databaseExtension.Status.Phase = "Failed"
		databaseExtension.Status.Message = err.Error()
		if updateErr := r.Status().Update(ctx, databaseExtension); updateErr != nil {
			logger.Error(updateErr)
			return reconcile.Result{}, updateErr
		}
		logger.Error(err)
		return reconcile.Result{}, err
	}

	databaseExtension.Status.Phase = "Applied"
	databaseExtension.Status.AppliedAt = time.Now().Unix()
	databaseExtension.Status.Message = "Extension successfully applied"

	if err := r.Status().Update(ctx, databaseExtension); err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}

	logger.Debug("extension successfully applied")
	return reconcile.Result{}, nil
}
