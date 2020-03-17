package table

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileTable) planConfigMap(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.ConfigMap, error) {
	b, err := yaml.Marshal(table.Spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal yaml spec")
	}

	tableData := make(map[string]string)
	tableData["table.yaml"] = string(b)

	name := table.Name
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: database.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		Data: tableData,
	}

	return configMap, nil
}

func (r *ReconcileTable) planPod(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
	imageName := "schemahero/schemahero:alpha"
	nodeSelector := make(map[string]string)
	driver := ""
	connectionURI := ""

	if database.Spec.SchemaHero != nil {
		if database.Spec.SchemaHero.Image != "" {
			imageName = database.Spec.SchemaHero.Image
		}

		nodeSelector = database.Spec.SchemaHero.NodeSelector
	}

	if database.Spec.Connection.Postgres != nil {
		driver = "postgres"
		uri, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.Postgres.URI)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read postgres connection uri")
		}
		connectionURI = uri
	} else if database.Spec.Connection.Mysql != nil {
		driver = "mysql"
		uri, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.Mysql.URI)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read mysql connection uri")
		}
		connectionURI = uri
	}

	if driver == "" {
		return nil, errors.New("unknown database driver")
	}

	labels := make(map[string]string)
	labels["schemahero-name"] = table.Name
	labels["schemahero-namespace"] = table.Namespace
	labels["schemahero-role"] = "plan"

	name := fmt.Sprintf("%s-plan", table.Name)
	configMapName := table.Name

	args := []string{
		"plan",
		"--driver",
		driver,
		"--uri",
		connectionURI,
		"--spec-file",
		"/specs/table.yaml",
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
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
					Args:            args,
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
								Name: configMapName,
							},
						},
					},
				},
			},
		},
	}

	return pod, nil
}

func (r *ReconcileTable) ensureTableConfigMap(desiredConfigMap *corev1.ConfigMap) error {
	existingConfigMap := corev1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: desiredConfigMap.Name, Namespace: desiredConfigMap.Namespace}, &existingConfigMap); err != nil {
		if kuberneteserrors.IsNotFound(err) {
			err = r.Create(context.Background(), desiredConfigMap)
			if err != nil {
				return errors.Wrap(err, "failed to create configmap")
			}
		}

		return errors.Wrap(err, "failed to get existing configmap")
	}

	return nil
}

func (r *ReconcileTable) ensureTablePod(desiredPod *corev1.Pod) error {
	existingPod := corev1.Pod{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: desiredPod.Name, Namespace: desiredPod.Namespace}, &existingPod); err != nil {
		if kuberneteserrors.IsNotFound(err) {
			err = r.Create(context.Background(), desiredPod)
			if err != nil {
				return errors.Wrap(err, "failed to create table migration pod")
			}

			return nil
		}

		return errors.Wrap(err, "failed to get existing pod object")
	}

	return nil
}
