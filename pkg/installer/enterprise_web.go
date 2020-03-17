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

func heroWebServiceYAML(namespace string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(heroWebService(namespace), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal hero web service")
	}

	return result.Bytes(), nil
}

func heroWebService(namespace string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hero-web",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "hero-web",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "hero-web",
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

func heroWebYAML(namespace string, tag string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(heroWeb(namespace, tag), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal hero web")
	}

	return result.Bytes(), nil
}

func heroWeb(namespace string, tag string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schemahero-web",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "hero-web",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "hero-web",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           fmt.Sprintf(`registry.replicated.com/schemahero-enterprise/hero-web:%s`, tag),
							ImagePullPolicy: corev1.PullAlways,
							Name:            "hero-web",
							Command: []string{
								"/bin/bash",
								"-c",
								"envsubst '$$ENV' < /etc/nginx/conf.d/default.template > /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENV",
									Value: "enterprise",
								},
							},
						},
					},
				},
			},
		},
	}
}
