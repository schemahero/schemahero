package database

import (
	"context"
	"fmt"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDatabase) ensurePostgresWatch(instance *databasesv1alpha3.Database) error {
	imageName := "schemahero/schemahero:alpha"
	nodeSelector := make(map[string]string)

	if instance.Spec.SchemaHero != nil {
		if instance.Spec.SchemaHero.Image != "" {
			imageName = instance.Spec.SchemaHero.Image
		}

		nodeSelector = instance.Spec.SchemaHero.NodeSelector
	}

	driver := "postgres"
	connectionURI, err := r.readConnectionURI(instance.Namespace, instance.Spec.Connection.Postgres.URI)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	// The pod created by this deployment will require access to the database object
	// defined in instance.Name. By default, RBAC will likely prevent this, so we need
	// to create a role and assign a service account to to pod
	if err := r.ensureWatchRBAC(instance); err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-watch",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "watch"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "watch"}},
				Spec: corev1.PodSpec{
					NodeSelector:       nodeSelector,
					ServiceAccountName: instance.Name,
					Containers: []corev1.Container{
						{
							Name:            "schemahero",
							Image:           imageName,
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"watch",
								"--driver",
								driver,
								"--uri",
								connectionURI,
								"--namespace",
								instance.Namespace,
								"--instance",
								instance.Name,
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.Background(), deploy)
		return err
	} else if err != nil {
		return err
	}

	// TODO diff and update!

	return nil
}
