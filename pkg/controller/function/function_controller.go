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

package function

import (
	"context"
	"slices"
	"time"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
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
	return &ReconcileFunction{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("function-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &schemasv1alpha4.Function{}, &handler.TypedEnqueueRequestForObject[*schemasv1alpha4.Function]{}))
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileFunction{}

type ReconcileFunction struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileFunction) dropFunction(ctx context.Context, function *schemasv1alpha4.Function) error {
	logger.Debug("dropping function",
		zap.String("name", function.Name),
		zap.String("namespace", function.Namespace))

	dbInstance, err := r.getDatabaseFromFunction(ctx, function)
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

	if driver != "postgres" || function.Spec.Schema.Postgres == nil {
		logger.Debug("doing nothing since the function is not configured for Postgres",
			zap.String("name", function.Name),
			zap.String("namespace", function.Namespace))
		return nil
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	// Get a connection to plan the function drop
	conn, err := db.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Mark function as deleted for drop operation
	functionSchema := &schemasv1alpha4.PostgresqlFunctionSchema{
		Schema:    function.Spec.Schema.Postgres.Schema,
		Lang:      function.Spec.Schema.Postgres.Lang,
		Params:    function.Spec.Schema.Postgres.Params,
		ReturnSet: function.Spec.Schema.Postgres.ReturnSet,
		Return:    function.Spec.Schema.Postgres.Return,
		As:        function.Spec.Schema.Postgres.As,
		IsDeleted: true, // Special marker for drop operation
	}

	statements, err := conn.PlanFunctionSchema(function.Spec.Name, functionSchema)
	if err != nil {
		return err
	}

	return db.ApplySync(statements)
}

func (r *ReconcileFunction) getDatabaseFromFunction(ctx context.Context, function *schemasv1alpha4.Function) (*databasesv1alpha4.Database, error) {
	database := &databasesv1alpha4.Database{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      function.Spec.Database,
		Namespace: function.Namespace,
	}, database)
	if err != nil {
		return nil, err
	}

	return database, nil
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=functions/status,verbs=get;update;patch
func (r *ReconcileFunction) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger.Debug("reconciling function",
		zap.String("kind", "function"),
		zap.String("name", request.Name),
		zap.String("namespace", request.Namespace))

	function := &schemasv1alpha4.Function{}
	err := r.Get(ctx, request.NamespacedName, function)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	finalizerName := "functions.schemas.schemahero.io/finalizer"

	if !function.ObjectMeta.DeletionTimestamp.IsZero() {
		if function.Spec.RemoveOnDeletion && slices.Contains(function.ObjectMeta.Finalizers, finalizerName) {
			if err = r.dropFunction(ctx, function); err != nil {
				return reconcile.Result{}, err
			}

			function.ObjectMeta.Finalizers = slices.DeleteFunc(function.ObjectMeta.Finalizers, func(s string) bool {
				return s == finalizerName
			})
			if err = r.Update(ctx, function); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	if function.Spec.RemoveOnDeletion && !slices.Contains(function.ObjectMeta.Finalizers, finalizerName) {
		function.ObjectMeta.Finalizers = append(function.ObjectMeta.Finalizers, finalizerName)
		if err = r.Update(ctx, function); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	dbInstance, err := r.getDatabaseFromFunction(ctx, function)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			logger.Debug("database not found, requeuing", zap.String("database", function.Spec.Database))
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

	if driver != "postgres" || function.Spec.Schema.Postgres == nil {
		logger.Debug("not a postgres database or no postgres function specified, skipping")
		return reconcile.Result{}, nil
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	// Get a connection to plan the function creation
	conn, err := db.GetConnection(ctx)
	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}
	defer conn.Close()

	// Use PlanFunctionSchema to create statements
	statements, err := conn.PlanFunctionSchema(function.Spec.Name, function.Spec.Schema.Postgres)
	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}

	if err := db.ApplySync(statements); err != nil {
		function.Status.Phase = "Failed"
		function.Status.Message = err.Error()
		if updateErr := r.Status().Update(ctx, function); updateErr != nil {
			logger.Error(updateErr)
			return reconcile.Result{}, updateErr
		}
		logger.Error(err)
		return reconcile.Result{}, err
	}

	function.Status.Phase = "Applied"
	function.Status.AppliedAt = time.Now().Unix()
	function.Status.Message = "Function successfully applied"

	if err := r.Status().Update(ctx, function); err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}

	logger.Debug("function successfully applied")
	return reconcile.Result{}, nil
}
