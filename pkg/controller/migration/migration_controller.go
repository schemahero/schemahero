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

package migration

import (
	"context"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Migration Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMigration{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("migration-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Migration
	err = c.Watch(&source.Kind{Type: &schemasv1alpha3.Migration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrap(err, "failed to start watch on migrations")
	}

	// Migrations are executed as pods, so we should also watch pod lifecycle
	generatedClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	generatedInformers := kubeinformers.NewSharedInformerFactory(generatedClient, time.Minute)
	err = mgr.Add(manager.RunnableFunc(func(s <-chan struct{}) error {
		generatedInformers.Start(s)
		<-s
		return nil
	}))
	if err != nil {
		return err
	}

	err = c.Watch(&source.Informer{
		Informer: generatedInformers.Core().V1().Pods().Informer(),
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrap(err, "failed to start watch on pods")
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMigration{}

// ReconcileMigration reconciles a Migration object
type ReconcileMigration struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Migration object and makes changes based on the state read
// and what is in the Migration.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=migrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=migrations/status,verbs=get;update;patch
func (r *ReconcileMigration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// This reconcile loop will be called for all Migration objects and all pods
	// because of the informer that we have set up
	// The behavior here is pretty different depending on the type
	// so this function is simply an entrypoint that executes the right reconcile loop
	instance, instanceErr := r.getInstance(request)
	if instanceErr == nil {
		result, err := r.reconcileMigration(context.Background(), instance)
		if err != nil {
			logger.Error(err)
		}
		return result, err
	}

	pod := &corev1.Pod{}
	podErr := r.Get(context.Background(), request.NamespacedName, pod)
	if podErr == nil {
		result, err := r.reconcilePod(context.Background(), pod)
		if err != nil {
			logger.Error(err)
		}
		return result, err
	}

	return reconcile.Result{}, errors.New("unknown error in migration reconciler")
}

func (r *ReconcileMigration) getInstance(request reconcile.Request) (*schemasv1alpha3.Migration, error) {
	v1alpha3instance := &schemasv1alpha3.Migration{}
	err := r.Get(context.Background(), request.NamespacedName, v1alpha3instance)
	if err != nil {
		return nil, err // don't wrap
	}

	return v1alpha3instance, nil
}

func (r *ReconcileMigration) readConnectionURI(database *databasesv1alpha3.Database) (string, error) {
	var valueOrValueFrom *databasesv1alpha3.ValueOrValueFrom

	if database.Spec.Connection.Postgres != nil {
		valueOrValueFrom = &database.Spec.Connection.Postgres.URI
	} else if database.Spec.Connection.Mysql != nil {
		valueOrValueFrom = &database.Spec.Connection.Mysql.URI
	} else if database.Spec.Connection.CockroachDB != nil {
		valueOrValueFrom = &database.Spec.Connection.CockroachDB.URI
	}

	if valueOrValueFrom == nil {
		return "", errors.New("cannnot get value from unknown type")
	}

	if valueOrValueFrom.Value != "" {
		return valueOrValueFrom.Value, nil
	}

	if valueOrValueFrom.ValueFrom == nil {
		return "", errors.New("value and valueFrom cannot both be nil/empty")
	}

	if valueOrValueFrom.ValueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		secretNamespacedName := types.NamespacedName{
			Name:      valueOrValueFrom.ValueFrom.SecretKeyRef.Name,
			Namespace: database.Namespace,
		}

		if err := r.Get(context.Background(), secretNamespacedName, secret); err != nil {
			if kuberneteserrors.IsNotFound(err) {
				return "", errors.New("table secret not found")
			} else {
				return "", errors.Wrap(err, "failed to get existing connection secret")
			}
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	if database.UsingVault() {
		// this feels wrong, but also doesn't make sense to return a
		// a URI ref as a connection URI?
		return "", nil
	}

	return "", errors.New("unable to find supported valueFrom")
}
