package table

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func configMapNameForPlan(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) string {
	shortID, err := getShortIDForTableSpec(table)
	if err != nil {
		return table.Name
	}
	configMapName := fmt.Sprintf("%s-%s-%s-plan", database.Name, table.Name, shortID)
	if len(apimachineryvalidation.NameIsDNSSubdomain(configMapName, false)) > 0 {
		configMapName = fmt.Sprintf("%s-%s-plan", table.Name, shortID)
		if len(apimachineryvalidation.NameIsDNSSubdomain(configMapName, false)) > 0 {
			configMapName = fmt.Sprintf("%s-plan", shortID)
		}
	}

	return configMapName
}

func podNameForPlan(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) string {
	shortID, err := getShortIDForTableSpec(table)
	if err != nil {
		return table.Name
	}
	podName := fmt.Sprintf("%s-%s-%s-plan", database.Name, table.Name, shortID)
	if len(apimachineryvalidation.NameIsDNSSubdomain(podName, false)) > 0 {
		podName = fmt.Sprintf("%s-%s-plan", table.Name, shortID)
		if len(apimachineryvalidation.NameIsDNSSubdomain(podName, false)) > 0 {
			podName = fmt.Sprintf("%s-plan", shortID)
		}
	}

	return podName
}

func getShortIDForTableSpec(table *schemasv1alpha3.Table) (string, error) {
	b, err := yaml.Marshal(table.Spec)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal yaml spec")
	}

	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)[:7], nil
}

func getPlanConfigMap(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.ConfigMap, error) {
	b, err := yaml.Marshal(table.Spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal yaml spec")
	}

	tableData := make(map[string]string)
	tableData["table.yaml"] = string(b)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapNameForPlan(database, table),
			Namespace: table.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		Data: tableData,
	}

	return configMap, nil
}

func getPlanServiceAccount(database *databasesv1alpha3.Database) *corev1.ServiceAccount {
	b := true
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      database.Name,
			Namespace: database.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		AutomountServiceAccountToken: &b,
	}

	return sa
}

func (r *ReconcileTable) getPlanPod(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
	imageName := "schemahero/schemahero:alpha"
	if os.Getenv("SCHEMAHERO_IMAGE_NAME") != "" {
		imageName = os.Getenv("SCHEMAHERO_IMAGE_NAME")
	}

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
	} else if database.Spec.Connection.CockroachDB != nil {
		driver = "cockroachdb"
		uri, err := r.readConnectionURI(database.Namespace, database.Spec.Connection.CockroachDB.URI)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read cockroachdb connection uri")
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

	args := []string{
		"plan",
		"--driver", driver,
		"--spec-file", "/specs/table.yaml",
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

	// Add serviceAccount
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podNameForPlan(database, table),
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
								Name: configMapNameForPlan(database, table),
							},
						},
					},
				},
			},
		},
	}

	return pod, nil
}
