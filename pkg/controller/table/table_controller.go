/*
Copyright 2019 Replicated, Inc.

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
	goerrors "errors"

	databasesv1alpha1 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha1"
	databasesclientv1alpha1 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha1"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Table Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTable{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("table-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Table
	err = c.Watch(&source.Kind{Type: &schemasv1alpha1.Table{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Table - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &schemasv1alpha1.Table{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTable{}

// ReconcileTable reconciles a Table object
type ReconcileTable struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Table object and makes changes based on the state read
// and what is in the Table.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables/status,verbs=get;update;patch
func (r *ReconcileTable) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Table instance
	instance := &schemasv1alpha1.Table{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	connection, err := r.getDatabaseConnection(instance.Namespace, instance.Spec.Database)
	if err != nil {
		return reconcile.Result{}, err
	}

	matchingType := r.checkDatabaseTypeMatches(connection, instance.Spec.Schema)
	if !matchingType {
		return reconcile.Result{}, goerrors.New("unable to deploy table to connection of different type")
	}

	if err := r.deploy(connection, instance.Spec); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTable) getDatabaseConnection(namespace string, name string) (*databasesv1alpha1.DatabaseConnection, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	databasesClient, err := databasesclientv1alpha1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	database, err := databasesClient.Databases(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &database.Connection, nil
}

func (r *ReconcileTable) checkDatabaseTypeMatches(connection *databasesv1alpha1.DatabaseConnection, tableSchema *schemasv1alpha1.TableSchema) bool {
	if connection.Postgres != nil {
		return tableSchema.Postgres != nil
	} else if connection.Mysql != nil {
		return tableSchema.Mysql != nil
	}

	return false
}

func (r *ReconcileTable) deploy(connection *databasesv1alpha1.DatabaseConnection, instanceSpec schemasv1alpha1.TableSpec) error {
	if connection.Postgres != nil {
		return r.deployPostgres(connection.Postgres, instanceSpec.Name, instanceSpec.Schema.Postgres)
	} else if connection.Mysql != nil {
		return r.deployMysql(connection.Mysql, instanceSpec.Name, instanceSpec.Schema.Mysql)
	}

	return goerrors.New("unknown database type")
}

func (r *ReconcileTable) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha1.ValueOrValueFrom) (string, error) {
	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	if valueOrValueFrom.ValueFrom == nil {
		return "", goerrors.New("value and valueFrom cannot both be nil/empty")
	}

	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Name:      valueOrValueFrom.ValueFrom.SecretKeyRef.Name,
			Namespace: namespace,
		}

		if err := r.Get(context.Background(), secretNamespacedName, secret); err != nil {
			if kuberneteserrors.IsNotFound(err) {
				return "", goerrors.New("secret not found")
			} else {
				return "", err
			}
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	return "", goerrors.New("unable to find supported valueFrom")
}
