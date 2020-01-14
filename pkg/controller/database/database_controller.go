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

	"github.com/pkg/errors"
	databasesv1alpha2 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
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

	// Watch for changes to Database kinds
	err = c.Watch(&source.Kind{Type: &databasesv1alpha2.Database{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &databasesv1alpha3.Database{}}, &handler.EnqueueRequestForObject{})
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
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases/status,verbs=get;update;patch
func (r *ReconcileDatabase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance, err := r.getInstance(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	// A "database" object is realized in the cluster as a deployment object,
	// in the namespace specified in the custom resource,

	logger.Debug("reconciling database",
		zap.String("name", instance.Name))

	// TODO add watches back in if we need them.  Today they aren't used for anything

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

func (r *ReconcileDatabase) getInstance(request reconcile.Request) (*databasesv1alpha3.Database, error) {
	v1alpha2instance := &databasesv1alpha2.Database{}
	err := r.Get(context.Background(), request.NamespacedName, v1alpha2instance)
	if err == nil {
		return databasesv1alpha3.ConvertFromV1Alpha2(v1alpha2instance), nil
	}

	instance := &databasesv1alpha3.Database{}
	err = r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databasesv1alpha3 instance")
	}

	return instance, nil
}

func (r *ReconcileDatabase) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha3.ValueOrValueFrom) (string, error) {
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
