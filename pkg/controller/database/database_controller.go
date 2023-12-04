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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ReconcileDatabase{}

// ReconcileDatabase reconciles a Database object
type ReconcileDatabase struct {
	client.Client
	scheme       *runtime.Scheme
	managerImage string
	managerTag   string
	debugLogs    bool
}

// Reconcile reads that state of the cluster for a Database object and makes changes based on the state read
// and what is in the Database.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases/status,verbs=get;update;patch
func (r *ReconcileDatabase) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	databaseInstance, err := r.getInstance(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	// A "database" object is realized in the cluster as a deployment object,
	// in the namespace specified in the custom resource,

	logger.Debug("reconciling database",
		zap.String("name", databaseInstance.Name))

	statefulsetName := fmt.Sprintf("%s-controller", databaseInstance.Name)

	// override default schemahero image for plans and applies, if present
	var schemaHeroManagerImage string
	if databaseInstance.Spec.SchemaHero != nil && databaseInstance.Spec.SchemaHero.Image != "" {
		schemaHeroManagerImage = databaseInstance.Spec.SchemaHero.Image
	} else {
		schemaHeroManagerImage = fmt.Sprintf("%s:%s", r.managerImage, r.managerTag)
	}

	vaultAnnotations, err := databaseInstance.GetVaultAnnotations()
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to get vault annotations"))
		return reconcile.Result{}, err
	}

	if vaultAnnotations == nil {
		vaultAnnotations = map[string]string{}
	}

	if err := r.reconcileRBAC(ctx, databaseInstance); err != nil {
		logger.Error(errors.Wrap(err, "failed to reconcile rbac"))
		return reconcile.Result{}, err
	}

	// TODO detect k8s version and use appsv1 or appsv1beta

	serviceAccountName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)
	labels := createLabels(databaseInstance)
	annotations := createAnnotations(databaseInstance)
	nodeSelectors := createNodeSelectors(databaseInstance)
	tolerations := createTolerations(databaseInstance)
	desiredStatefulSet := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        statefulsetName,
			Namespace:   databaseInstance.Namespace,
			Labels:      *labels,
			Annotations: *annotations,
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
					Labels:      *labels,
					Annotations: vaultAnnotations,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"linux",
												},
											},
										},
									},
								},
							},
						},
					},
					NodeSelector:                  nodeSelectors,
					Tolerations:                   tolerations,
					TerminationGracePeriodSeconds: &tenSeconds,
					ServiceAccountName:            serviceAccountName,
					Containers: []corev1.Container{
						{
							Image:           schemaHeroManagerImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
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
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	if r.debugLogs {
		desiredStatefulSet.Spec.Template.Spec.Containers[0].Args = append(desiredStatefulSet.Spec.Template.Spec.Containers[0].Args, "--log-level", "debug")
	}

	existingStatefulset := appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: databaseInstance.Namespace,
		Name:      statefulsetName,
	}, &existingStatefulset)

	if err == nil {
		existingStatefulset.Spec = *desiredStatefulSet.Spec.DeepCopy()
		if err := r.Update(ctx, &desiredStatefulSet); err != nil {
			logger.Error(err)
			return reconcile.Result{}, err
		}
	} else if kuberneteserrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(databaseInstance, &desiredStatefulSet, r.scheme); err != nil {
			logger.Error(err)
			return reconcile.Result{}, err
		}

		if err := r.Create(ctx, &desiredStatefulSet); err != nil {
			logger.Error(err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		logger.Error(errors.Wrapf(err, "failed to get statefulset %s", statefulsetName))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func createLabels(db *databasesv1alpha4.Database) *map[string]string {
	l := map[string]string{
		"control-plane": "schemahero",
		"database":      db.Name,
	}

	if db.Spec.Template != nil {
		for k, v := range db.Spec.Template.ObjectMeta.Labels {
			l[k] = v
		}
	}

	return &l
}

func createAnnotations(db *databasesv1alpha4.Database) *map[string]string {
	a := map[string]string{}

	if db.Spec.Template != nil {
		for k, v := range db.Spec.Template.ObjectMeta.Annotations {
			a[k] = v
		}
	}

	return &a
}

func createNodeSelectors(db *databasesv1alpha4.Database) map[string]string {
	a := map[string]string{}

	if db.Spec.SchemaHero != nil {
		for k, v := range db.Spec.SchemaHero.NodeSelector {
			a[k] = v
		}
	}

	return a
}

func createTolerations(db *databasesv1alpha4.Database) []corev1.Toleration {
	a := []corev1.Toleration{}

	if db.Spec.SchemaHero != nil {
		for k, v := range db.Spec.SchemaHero.Tolerations {
			c := corev1.Toleration{
				Effect:   corev1.TaintEffect(v.Effect),
				Key:      v.Key,
				Operator: corev1.TolerationOperator(v.Operator),
				Value:    v.Value,
			}
			a[k] = c
		}
	}

	return a
}

func (r *ReconcileDatabase) getInstance(request reconcile.Request) (*databasesv1alpha4.Database, error) {
	instance := &databasesv1alpha4.Database{}
	err := r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databasesv1alpha4 instance")
	}

	return instance, nil
}
