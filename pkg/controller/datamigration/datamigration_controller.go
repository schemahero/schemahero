/*
Copyright 2025 The SchemaHero Authors

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

package datamigration

import (
	"context"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new DataMigration Controller and adds it to the Manager with default RBAC.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDataMigration{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("datamigration-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to DataMigration
	err = c.Watch(source.Kind(mgr.GetCache(), &schemasv1alpha4.DataMigration{}, &handler.TypedEnqueueRequestForObject[*schemasv1alpha4.DataMigration]{}))
	if err != nil {
		return errors.Wrap(err, "failed to start watch on datamigrations")
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDataMigration{}

// ReconcileDataMigration reconciles a DataMigration object
type ReconcileDataMigration struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads the state of the cluster for a DataMigration object and creates a Migration
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=datamigrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=datamigrations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=migrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch
func (r *ReconcileDataMigration) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger.Debug("reconciling datamigration",
		zap.String("name", request.Name),
		zap.String("namespace", request.Namespace))

	// Fetch the DataMigration instance
	instance := &schemasv1alpha4.DataMigration{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		// Error reading the object - requeue the request
		return reconcile.Result{}, err
	}

	// Skip if already processed
	if instance.Status.MigrationName != "" {
		logger.Debug("datamigration already has associated migration",
			zap.String("migration", instance.Status.MigrationName))
		return reconcile.Result{}, nil
	}

	// Reconcile the DataMigration
	return r.reconcileDataMigration(ctx, instance)
}