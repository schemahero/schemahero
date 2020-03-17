package installer

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func heroAPIServiceYAML(namespace string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(heroAPIService(namespace), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal hero api service")
	}

	return result.Bytes(), nil
}

func heroAPIService(namespace string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hero-api",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "hero-api",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "hero-api",
			},
			Ports: []corev1.ServicePort{
				{
					Port: 3000,
					Name: "http",
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func heroAPIYAML(namespace string, tag string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(heroAPI(namespace, tag), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal hero api")
	}

	return result.Bytes(), nil
}

func heroAPI(namespace string, tag string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schemahero-api",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "hero-api",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "hero-api",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           fmt.Sprintf(`repl{{ LocalImageName "registry.replicated.com/schemahero-enterprise/hero-api:%s"}}`, tag),
							ImagePullPolicy: corev1.PullAlways,
							Name:            "hero-api",
							Command: []string{
								"/go/src/github.com/schemahero/hero-web/api/bin/hero-api",
								"run",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "FRONTEND_URL",
									Value: "http://localhost:8888",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
									Name:          "http",
								},
							},
						},
					},
				},
			},
		},
	}
}
