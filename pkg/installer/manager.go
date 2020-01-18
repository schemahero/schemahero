package installer

import (
	"bytes"

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

func namespaceYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(namespace(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal namespace")
	}

	return result.Bytes(), nil
}

func ensureNamespace(clientset *kubernetes.Clientset) error {
	_, err := clientset.CoreV1().Namespaces().Get("schemahero-system", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get namespace")
		}

		_, err := clientset.CoreV1().Namespaces().Create(namespace())
		if err != nil {
			return errors.Wrap(err, "failed to create namespace")
		}
	}

	return nil
}

func namespace() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "schemahero-system",
		},
	}
}

func serviceYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(service(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal service")
	}

	return result.Bytes(), nil
}

func ensureService(clientset *kubernetes.Clientset) error {
	_, err := clientset.CoreV1().Services("schemahero-system").Get("controller-manager-service", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get service")
		}

		_, err := clientset.CoreV1().Secrets("schemahero-system").Create(secret())
		if err != nil {
			return errors.Wrap(err, "failed to create service")
		}
	}

	return nil
}

func service() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-webhook-server",
			Namespace: "schemahero-system",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"control-plane": "schemahero",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					TargetPort: intstr.FromInt(9443),
				},
			},
		},
	}
}

func secretYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(secret(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal secret")
	}

	return result.Bytes(), nil
}

func ensureSecret(clientset *kubernetes.Clientset) error {
	_, err := clientset.CoreV1().Secrets("schemahero-system").Get("webhook-server-secret", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get secret")
		}

		_, err := clientset.CoreV1().Secrets("schemahero-system").Create(secret())
		if err != nil {
			return errors.Wrap(err, "failed to create secret")
		}
	}

	return nil
}

func secret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webhook-server-secret",
			Namespace: "schemahero-system",
		},
	}
}

func managerYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(manager(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal manager")
	}

	return result.Bytes(), nil
}

func ensureManager(clientset *kubernetes.Clientset) error {
	_, err := clientset.AppsV1().StatefulSets("schemahero-system").Get("schemahero", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get statefulset")
		}

		_, err := clientset.AppsV1().StatefulSets("schemahero-system").Create(manager())
		if err != nil {
			return errors.Wrap(err, "failed to create statefulset")
		}
	}

	return nil
}

func manager() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schemahero",
			Namespace: "schemahero-system",
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
							Image:           "schemahero/schemahero-manager:0.8.0-alpha",
							ImagePullPolicy: corev1.PullAlways,
							Name:            "manager",
							Command:         []string{"/manager"},
							Env: []corev1.EnvVar{
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
