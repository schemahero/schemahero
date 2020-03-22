package migration

import (
	"fmt"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getApplyConfigMap(migrationID string, namespace string, preparedStatement string) (*corev1.ConfigMap, error) {
	data := make(map[string]string)
	data["ddl.sql"] = preparedStatement

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      migrationID,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		Data: data,
	}

	return configMap, nil
}

func getApplyPod(migrationID string, namespace string, connectionURI string, driver string, database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
	imageName := "schemahero/schemahero:alpha"
	nodeSelector := make(map[string]string)

	if database.Spec.SchemaHero != nil {
		if database.Spec.SchemaHero.Image != "" {
			imageName = database.Spec.SchemaHero.Image
		}

		nodeSelector = database.Spec.SchemaHero.NodeSelector
	}

	labels := make(map[string]string)
	labels["schemahero-name"] = table.Name
	labels["schemahero-namespace"] = table.Namespace
	labels["schemahero-role"] = "apply"

	name := fmt.Sprintf("%s-apply", table.Name)

	args := []string{
		"apply",
		"--driver",
		driver,
		"--uri",
		connectionURI,
		"--ddl",
		"/input/ddl.sql",
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
							Name:      "input",
							MountPath: "/input",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "input",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: migrationID,
							},
						},
					},
				},
			},
		},
	}

	return pod, nil
}
