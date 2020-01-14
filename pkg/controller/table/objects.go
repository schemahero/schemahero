package table

import (
	"fmt"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileTable) applyConfigMap(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table, plan *schemasv1alpha3.TablePlan) (*corev1.ConfigMap, error) {
	tableData := make(map[string]string)
	tableData["ddl.sql"] = string(plan.DDL)

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
	if err := controllerutil.SetControllerReference(table, configMap, r.scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller reference")
	}

	return configMap, nil
}

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
	if err := controllerutil.SetControllerReference(table, configMap, r.scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller reference")
	}

	return configMap, nil
}

func (r *ReconcileTable) applyPod(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
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
			return nil, errors.Wrap(err, "failed to read postgres connection uri")
		}
		connectionURI = uri
	} else if database.Connection.Mysql != nil {
		driver = "mysql"
		uri, err := r.readConnectionURI(database.Namespace, database.Connection.Mysql.URI)
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
	labels["schemahero-role"] = "apply"

	name := fmt.Sprintf("%s-apply", table.Name)
	configMapName := table.Name

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
								Name: configMapName,
							},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(table, pod, r.scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller ref")
	}

	return pod, nil
}

func (r *ReconcileTable) planPod(database *databasesv1alpha3.Database, table *schemasv1alpha3.Table) (*corev1.Pod, error) {
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
			return nil, errors.Wrap(err, "failed to read postgres connection uri")
		}
		connectionURI = uri
	} else if database.Connection.Mysql != nil {
		driver = "mysql"
		uri, err := r.readConnectionURI(database.Namespace, database.Connection.Mysql.URI)
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
	if err := controllerutil.SetControllerReference(table, pod, r.scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller ref")
	}

	return pod, nil
}
