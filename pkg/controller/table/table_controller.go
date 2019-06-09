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
	"fmt"
	"time"

	databasesv1alpha2 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	databasesclientv1alpha2 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha2"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	err = c.Watch(&source.Kind{Type: &schemasv1alpha2.Table{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Add an informer on pods, which are created to deploy schemas. the informer will
	// update the status of the table custom resource and do a little garbage collection
	generatedClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	generatedInformers := kubeinformers.NewSharedInformerFactory(generatedClient, time.Second)
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
		return err
	}

	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &schemasv1alpha2.Table{},
	// })
	// if err != nil {
	// 	return err
	// }

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
	instance := &schemasv1alpha2.Table{}
	instanceErr := r.Get(context.TODO(), request.NamespacedName, instance)

	pod := &corev1.Pod{}
	podErr := r.Get(context.TODO(), request.NamespacedName, pod)

	// operator reconciler (table object)
	if instanceErr == nil {
		database, err := r.getDatabaseSpec(instance.Namespace, instance.Spec.Database)
		if err != nil {
			return reconcile.Result{}, err
		}
		if database == nil {
			// TODO get a real logger, this isn't a warning, it's expected, probably debug level
			fmt.Printf("requeuing table deployment for %q, database is not available\n", instance.Spec.Name)
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 10,
			}, nil
		}

		matchingType := r.checkDatabaseTypeMatches(&database.Connection, instance.Spec.Schema)
		if !matchingType {
			return reconcile.Result{}, goerrors.New("unable to deploy table to connection of different type")
		}

		if err := r.deploy(database, instance); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// pod informer
	if podErr == nil {
		podLabels := pod.GetObjectMeta().GetLabels()
		role, ok := podLabels["schemahero-role"]
		if !ok {
			return reconcile.Result{}, nil
		}

		if role != "table" {
			return reconcile.Result{}, nil
		}
		if pod.Status.Phase == corev1.PodSucceeded {
			// TODO: Update the status on the table object

			// Delete the pod and config map
			if err := r.Delete(context.TODO(), pod); err != nil {
				return reconcile.Result{}, err
			}

			// read the name of the config map
			configMapName := ""
			for _, volume := range pod.Spec.Volumes {
				if volume.Name == "specs" && volume.ConfigMap != nil {
					configMapName = volume.ConfigMap.Name
				}
			}

			configMap := corev1.ConfigMap{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: pod.Namespace}, &configMap)
			if err != nil {
				return reconcile.Result{}, err
			}

			if err := r.Delete(context.TODO(), &configMap); err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	}

	if errors.IsNotFound(instanceErr) {
		// Object not found, return.  Created objects are automatically garbage collected.
		// For additional cleanup logic use finalizers.
		return reconcile.Result{}, nil
	}
	// Error reading the object - requeue the request.
	return reconcile.Result{}, instanceErr
}

func (r *ReconcileTable) getDatabaseSpec(namespace string, name string) (*databasesv1alpha2.Database, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	databasesClient, err := databasesclientv1alpha2.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	database, err := databasesClient.Databases(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// tables might be deployed before a database... if this is the case
		// we don't want to crash, we want to re-reconcile later
		if kuberneteserrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// try to parse the secret too, the database may be deployed, but that doesn't mean it's ready
	// TODO this would be better as a status field on the database object, instead of this leaky
	// interface
	if database.Connection.Postgres != nil {
		_, err := r.readConnectionURI(database.Namespace, database.Connection.Postgres.URI)
		if err != nil {
			return nil, nil
		}
	} else if database.Connection.Mysql != nil {
		_, err := r.readConnectionURI(database.Namespace, database.Connection.Mysql.URI)
		if err != nil {
			return nil, nil
		}
	}

	return database, nil
}

func (r *ReconcileTable) checkDatabaseTypeMatches(connection *databasesv1alpha2.DatabaseConnection, tableSchema *schemasv1alpha2.TableSchema) bool {
	if connection.Postgres != nil {
		return tableSchema.Postgres != nil
	} else if connection.Mysql != nil {
		return tableSchema.Mysql != nil
	}

	return false
}

func (r *ReconcileTable) deploy(database *databasesv1alpha2.Database, table *schemasv1alpha2.Table) error {
	if err := r.ensureTableConfigMap(database, table); err != nil {
		return err
	}

	if err := r.ensureTablePod(database, table); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileTable) ensureTableConfigMap(database *databasesv1alpha2.Database, table *schemasv1alpha2.Table) error {
	b, err := yaml.Marshal(table.Spec)
	if err != nil {
		return err
	}

	tableData := make(map[string]string)
	tableData["table.yaml"] = string(b)

	existingConfigMap := corev1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: table.Name, Namespace: database.Namespace}, &existingConfigMap); err != nil {
		if kuberneteserrors.IsNotFound(err) {
			configMap := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      table.Name,
					Namespace: database.Namespace,
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: tableData,
			}
			if err := controllerutil.SetControllerReference(table, &configMap, r.scheme); err != nil {
				return err
			}
			err = r.Create(context.TODO(), &configMap)
			if err != nil {
				return err
			}
		}

		return err
	}

	return nil
}

func (r *ReconcileTable) ensureTablePod(database *databasesv1alpha2.Database, table *schemasv1alpha2.Table) error {
	imageName := "schemahero/schemahero:alpha"
	nodeSelector := make(map[string]string)
	driver := ""
	connectionURI := ""

	if database.SchemaHero != nil {
		if database.SchemaHero.Image != "" {
			imageName = database.SchemaHero.Image
		}

		nodeSelector = database.SchemaHero.NodeSelector
	}

	if database.Connection.Postgres != nil {
		driver = "postgres"
		uri, err := r.readConnectionURI(database.Namespace, database.Connection.Postgres.URI)
		if err != nil {
			return err
		}
		connectionURI = uri
	} else if database.Connection.Mysql != nil {
		driver = "mysql"
		uri, err := r.readConnectionURI(database.Namespace, database.Connection.Mysql.URI)
		if err != nil {
			return err
		}
		connectionURI = uri
	}

	if driver == "" {
		return goerrors.New("unknown database driver")
	}

	labels := make(map[string]string)
	labels["schemahero-role"] = "table"

	existingPod := corev1.Pod{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s-apply", table.Name), Namespace: database.Namespace}, &existingPod); err != nil {
		if kuberneteserrors.IsNotFound(err) {
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-apply", table.Name),
					Namespace: database.Namespace,
					Labels:    labels,
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				Spec: corev1.PodSpec{
					NodeSelector:       nodeSelector,
					ServiceAccountName: database.Name,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Image:           imageName,
							ImagePullPolicy: corev1.PullAlways,
							Name:            table.Name,
							Args: []string{
								"apply",
								"--driver",
								driver,
								"--uri",
								connectionURI,
								"--spec-file",
								"/specs/table.yaml",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "specs",
									MountPath: "/specs",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "specs",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: table.Name,
									},
								},
							},
						},
					},
				},
			}
			if err := controllerutil.SetControllerReference(table, &pod, r.scheme); err != nil {
				return err
			}
			err := r.Create(context.TODO(), &pod)
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func (r *ReconcileTable) readConnectionURI(namespace string, valueOrValueFrom databasesv1alpha2.ValueOrValueFrom) (string, error) {
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
				return "", goerrors.New("table secret not found")
			} else {
				return "", err
			}
		}

		return string(secret.Data[valueOrValueFrom.ValueFrom.SecretKeyRef.Key]), nil
	}

	return "", goerrors.New("unable to find supported valueFrom")
}
