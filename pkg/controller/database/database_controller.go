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

package database

import (
	"context"
	goerrors "errors"
	"fmt"

	databasesv1alpha2 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// Add creates a new Database Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDatabase{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("database-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Database
	err = c.Watch(&source.Kind{Type: &databasesv1alpha2.Database{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Database - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &databasesv1alpha2.Database{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDatabase{}

// ReconcileDatabase reconciles a Database object
type ReconcileDatabase struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Database object and makes changes based on the state read
// and what is in the Database.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases/status,verbs=get;update;patch
func (r *ReconcileDatabase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Database instance
	instance := &databasesv1alpha2.Database{}
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

	// If a gitops configuration is set on the database object, then
	// make sure that is creates and initialized

	if instance.GitOps != nil {
		if instance.GitOps.IsPlanEnabled {
			if err := r.ensureGitOpsPlan(instance); err != nil {
				gitopsPlanLoop = nil
				return reconcile.Result{}, err
			}
		}

		// if err := r.ensureGitOps(instance); err != nil {
		// 	gitopsLoop = nil
		// 	return reconcile.Result{}, err
		// }
	}

	// A "database" object is realized in the cluster as a deployment object,
	// in the namespace specified in the custom resource,

	if instance.Connection.Postgres != nil {
		if err = r.ensurePostgresWatch(instance); err != nil {
			return reconcile.Result{}, err
		}
	} else if instance.Connection.Mysql != nil {
		if err := r.ensureMysqlWatch(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDatabase) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha2.ValueOrValueFrom) (string, error) {
	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	if valueOrValueFrom.ValueFrom == nil {
		return "", goerrors.New("value and valueFrom cannot both be nil/empty")
	}

	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret := corev1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Name:      valueOrValueFrom.ValueFrom.SecretKeyRef.Name,
			Namespace: namespace,
		}

		if err := r.Get(context.TODO(), secretNamespacedName, &secret); err != nil {
			if kuberneteserrors.IsNotFound(err) {
				return "", fmt.Errorf("database secret (%s/%s) not found", secretNamespacedName.Namespace, secretNamespacedName.Name)
			} else {
				return "", err
			}
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	return "", goerrors.New("unable to find supported valueFrom")
}
