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

package table

import (
	"context"
	"time"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Table Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, databaseNames []string) error {
	return add(mgr, newReconciler(databaseNames, mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(databaseNames []string, mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTable{
		Client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		databaseNames: databaseNames,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("table-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Table
	err = c.Watch(source.Kind(mgr.GetCache(), &schemasv1alpha4.Table{}), &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrap(err, "failed to start watch on tables")
	}

	// Add an informer on pods, which are created to deploy schemas. the informer will
	// update the status of the table custom resource and do a little garbage collection
	generatedClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	generatedInformers := kubeinformers.NewSharedInformerFactory(generatedClient, time.Minute)
	err = mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		s := make(chan struct{})
		generatedInformers.Start(s)
		<-s
		return nil
	}))

	return err
}

var _ reconcile.Reconciler = &ReconcileTable{}

// ReconcileTable reconciles a Table object
type ReconcileTable struct {
	client.Client
	scheme        *runtime.Scheme
	databaseNames []string
}

// Reconcile reads that state of the cluster for a Table object and makes changes based on the state read
// and what is in the Table.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables/status,verbs=get;update;patch
func (r *ReconcileTable) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// This reconcile loop will be called for all Table objects and all pods
	// because of the informer that we have set up
	// The behavior here is pretty different depending on the type
	// so this function is simply an entrypoint that executes the right reconcile loop
	instance, err := r.getInstance(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	isThisController, err := r.isTableManagedByThisController(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !isThisController {
		logger.Debug("table instance is not managed by this controller",
			zap.String("table", instance.Name),
			zap.Strings("databaseNames", r.databaseNames))
		return reconcile.Result{}, nil
	}

	result, err := r.reconcileTable(ctx, instance)
	if err != nil {
		logger.Error(err)
	}

	return result, err
}

func (r *ReconcileTable) isTableManagedByThisController(instance *schemasv1alpha4.Table) (bool, error) {
	databaseName := instance.Spec.Database

	for _, managedDatabaseName := range r.databaseNames {
		if managedDatabaseName == databaseName {
			return true, nil
		}

		if managedDatabaseName == "*" {
			return true, nil
		}
	}

	return false, nil
}
