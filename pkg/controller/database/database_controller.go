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
	"fmt"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var tenSeconds = int64(10)
var defaultMode = int32(420)

// Add creates a new Database Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, managerImage string, managerTag string) error {
	return add(mgr, newReconciler(mgr, managerImage, managerTag))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, managerImage string, managerTag string) reconcile.Reconciler {
	return &ReconcileDatabase{
		Client:       mgr.GetClient(),
		scheme:       mgr.GetScheme(),
		managerImage: managerImage,
		managerTag:   managerTag,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	logger.Debug("adding database controller to manager")

	// Create a new controller
	c, err := controller.New("database-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Database kinds
	err = c.Watch(&source.Kind{
		Type: &databasesv1alpha4.Database{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDatabase{}

// ReconcileDatabase reconciles a Database object
type ReconcileDatabase struct {
	client.Client
	scheme       *runtime.Scheme
	managerImage string
	managerTag   string
}

// Reconcile reads that state of the cluster for a Database object and makes changes based on the state read
// and what is in the Database.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases/status,verbs=get;update;patch
func (r *ReconcileDatabase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	databaseInstance, err := r.getInstance(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	ctx := context.Background()

	// A "database" object is realized in the cluster as a deployment object,
	// in the namespace specified in the custom resource,

	logger.Debug("reconciling database",
		zap.String("name", databaseInstance.Name))

	statefulsetName := fmt.Sprintf("%s-controller", databaseInstance.Name)
	schemaHeroManagerImage := fmt.Sprintf("%s:%s", r.managerImage, r.managerTag)

	vaultAnnotations, err := databaseInstance.GetVaultAnnotations()

	if err != nil {
		logger.Error(err)
		return reconcile.Result{}, err
	}
	if vaultAnnotations == nil {
		vaultAnnotations = map[string]string{}
	}

	if err := r.reconcileRBAC(ctx, databaseInstance); err != nil {
		logger.Error(err)
		return reconcile.Result{}, nil
	}

	// TODO detect k8s version and use appsv1 or appsv1beta

	existingStatefulset := appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: databaseInstance.Namespace,
		Name:      statefulsetName,
	}, &existingStatefulset)
	if kuberneteserrors.IsNotFound(err) {
		// create
		statefulSet := appsv1.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      statefulsetName,
				Namespace: databaseInstance.Namespace,
				Labels: map[string]string{
					"control-plane": "schemahero",
					"database":      databaseInstance.Name,
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"control-plane": "schemahero",
						"database":      databaseInstance.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"control-plane": "schemahero",
							"database":      databaseInstance.Name,
						},
						Annotations: vaultAnnotations,
					},
					Spec: corev1.PodSpec{
						TerminationGracePeriodSeconds: &tenSeconds,
						ServiceAccountName:            "schemahero",
						Containers: []corev1.Container{
							{
								Image:           schemaHeroManagerImage,
								ImagePullPolicy: corev1.PullAlways,
								Name:            "manager",
								Command:         []string{"/manager"},
								Args: []string{
									"run",
									"--namespace", databaseInstance.Namespace,
									"--database-name", databaseInstance.Name,
								},
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1"),
										corev1.ResourceMemory: resource.MustParse("150Mi"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("50Mi"),
									},
								},
							},
						},
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(databaseInstance, &statefulSet, r.scheme); err != nil {
			logger.Error(err)
			return reconcile.Result{}, err
		}

		if err := r.Create(ctx, &statefulSet); err != nil {
			logger.Error(err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		logger.Error(err)
		return reconcile.Result{}, nil
	} else {
		// update with the new database details

		logger.Error(errors.New("updating table reconciler is not implemeted"))
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDatabase) getInstance(request reconcile.Request) (*databasesv1alpha4.Database, error) {
	instance := &databasesv1alpha4.Database{}
	err := r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databasesv1alpha4 instance")
	}

	return instance, nil
}
