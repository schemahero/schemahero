package migration

import (
	"fmt"
	"os"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func configMapNameForMigration(databaseName string, tableName string, migrationID string) string {
	configMapName := fmt.Sprintf("%s-%s-%s", databaseName, tableName, migrationID)
	if len(apimachineryvalidation.NameIsDNSSubdomain(configMapName, false)) > 0 {
		configMapName = fmt.Sprintf("%s-%s", tableName, migrationID)
		if len(apimachineryvalidation.NameIsDNSSubdomain(configMapName, false)) > 0 {
			configMapName = migrationID
		}
	}

	return configMapName
}

func getApplyConfigMap(migrationID string, namespace string, preparedStatement string, tableName string, databaseName string) (*corev1.ConfigMap, error) {
	data := make(map[string]string)
	data["ddl.sql"] = preparedStatement

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapNameForMigration(databaseName, tableName, migrationID),
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

func podNameForMigrationApply(databaseName string, tableName string, migrationID string) string {
	podName := fmt.Sprintf("%s-%s-%s-apply", databaseName, tableName, migrationID)
	if len(apimachineryvalidation.NameIsDNSSubdomain(podName, false)) > 0 {
		podName = fmt.Sprintf("%s-%s-apply", tableName, migrationID)
		if len(apimachineryvalidation.NameIsDNSSubdomain(podName, false)) > 0 {
			podName = fmt.Sprintf("%s-apply", migrationID)
		}
	}

	return podName
}

func getApplyPod(migrationID string, namespace string, connectionURI string, database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
	imageName := "schemahero/schemahero:alpha"
	if os.Getenv("SCHEMAHERO_IMAGE_NAME") != "" {
		imageName = os.Getenv("SCHEMAHERO_IMAGE_NAME")
	}

	nodeSelector := make(map[string]string)

	if database.Spec.SchemaHero != nil {
		if database.Spec.SchemaHero.Image != "" {
			imageName = database.Spec.SchemaHero.Image
		}

		nodeSelector = database.Spec.SchemaHero.NodeSelector
	}

	labels := make(map[string]string)
	labels["schemahero-name"] = migrationID
	labels["schemahero-namespace"] = table.Namespace
	labels["schemahero-role"] = "apply"

	driver := ""
	if database.Spec.Connection.Postgres != nil {
		driver = "postgres"
	} else if database.Spec.Connection.Mysql != nil {
		driver = "mysql"
	} else if database.Spec.Connection.CockroachDB != nil {
		driver = "cockroachdb"
	}

	args := []string{
		"apply",
		"--driver", driver,
		"--ddl", "/input/ddl.sql",
	}

	var vaultAnnotations map[string]string
	if database.UsingVault() {
		a, err := database.GetVaultAnnotations()
		if err != nil {
			return nil, err
		}
		vaultAnnotations = a

		args = append(args, "--vault-uri-ref", "/vault/secrets/schemaherouri")
	} else {
		args = append(args, "--uri", connectionURI)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podNameForMigrationApply(database.Name, table.Name, migrationID),
			Namespace:   database.Namespace,
			Labels:      labels,
			Annotations: vaultAnnotations,
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
								Name: configMapNameForMigration(database.Name, table.Name, migrationID),
							},
						},
					},
				},
			},
		},
	}

	return pod, nil
}
