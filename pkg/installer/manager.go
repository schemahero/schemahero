package installer

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

var tenSeconds = int64(10)
var defaultMode = int32(420)

func namespaceYAML(name string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(namespace(name), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal namespace")
	}

	return result.Bytes(), nil
}

func ensureNamespace(ctx context.Context, clientset *kubernetes.Clientset, name string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get namespace")
		}

		_, err := clientset.CoreV1().Namespaces().Create(ctx, namespace(name), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create namespace")
		}
	}

	return nil
}

func namespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func serviceYAML(namespace string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(service(namespace), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal service")
	}

	return result.Bytes(), nil
}

func ensureService(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Services(namespace).Get(ctx, "controller-manager-service", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get service")
		}

		_, err := clientset.CoreV1().Services(namespace).Create(ctx, service(namespace), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create service")
		}
	}

	return nil
}

func service(namespace string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-manager-service",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"control-plane": "schemahero",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					TargetPort: intstr.FromInt(9876),
				},
			},
		},
	}
}

func secretYAML(namespace string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(secret(namespace), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal secret")
	}

	return result.Bytes(), nil
}

func ensureSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Secrets(namespace).Get(ctx, "webhook-server-secret", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get secret")
		}

		_, err := clientset.CoreV1().Secrets(namespace).Create(ctx, secret(namespace), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create secret")
		}
	}

	return nil
}

func secret(namespace string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-server-secret",
			Namespace: namespace,
		},
	}
}

func managerYAML(namespace string, managerImage string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(manager(namespace, managerImage), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal manager")
	}

	return result.Bytes(), nil
}

func ensureManager(ctx context.Context, clientset *kubernetes.Clientset, namespace string, managerImage string) (bool, error) {
	existingManager, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, "schemahero", metav1.GetOptions{})

	if err != nil && !kuberneteserrors.IsNotFound(err) {
		return false, errors.Wrap(err, "get manager statefulset")
	}

	if kuberneteserrors.IsNotFound(err) {
		_, err := clientset.AppsV1().StatefulSets(namespace).Create(ctx, manager(namespace, managerImage), metav1.CreateOptions{})
		if err != nil {
			return false, errors.Wrap(err, "create manager statefulset")
		}

		return false, nil
	}

	// update the existing manager, but it's a statefulset, so
	// we can only mutate some fields
	existingManager.Spec = manager(namespace, managerImage).Spec

	_, err = clientset.AppsV1().StatefulSets(namespace).Update(ctx, existingManager, metav1.UpdateOptions{})
	if err != nil {
		return false, errors.Wrap(err, "update manager statefulset")
	}

	return true, nil
}

func manager(namespace string, managerImage string) *appsv1.StatefulSet {
	env := []corev1.EnvVar{
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "SECRET_NAME",
			Value: "webhook-server-secret",
		},
	}

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schemahero",
			Namespace: namespace,
			Labels: map[string]string{
				"control-plane": "schemahero",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"control-plane": "schemahero",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"control-plane": "schemahero",
					},
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
											{
												Key:      "kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"amd64",
												},
											},
										},
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &tenSeconds,
					Volumes: []corev1.Volume{
						{
							Name: "cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									DefaultMode: &defaultMode,
									SecretName:  "webhook-server-secret",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Image:           managerImage,
							ImagePullPolicy: corev1.PullAlways,
							Name:            "manager",
							Command:         []string{"/manager", "run", "--enable-database-controller"},
							Env:             env,
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
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9876,
									Name:          "webhook-server",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cert",
									MountPath: "/tmp/cert",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}
}
