package installer

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	extensionsv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
)

func migrationsCRDYAML(useExtensionsv1beta1 bool) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if useExtensionsv1beta1 {
		if err := s.Encode(mnigrationsCRDV1Beta1(), &result); err != nil {
			return nil, errors.Wrap(err, "failed to marshal migrations v1beta1 crd")
		}

	} else {
		if err := s.Encode(migrationsCRDV1(), &result); err != nil {
			return nil, errors.Wrap(err, "failed to marshal migrations v1 crd")
		}

	}
	return result.Bytes(), nil
}

func ensureMigrationsCRD(cfg *rest.Config, useExtensionsv1beta1 bool) error {
	if useExtensionsv1beta1 {
		extensionsClient, err := extensionsv1beta1client.NewForConfig(cfg)
		if err != nil {
			return errors.Wrap(err, "faild to create extensions client")
		}

		_, err := extensionsClient.CustomResourceDefinitions().Get("migrations.schemas.schemahero.io", metav1.GetOptions{})
		if err != nil {
			if !kuberneteserrors.IsNotFound(err) {
				return errors.Wrap(err, "failed to get migrations crd")
			}

			_, err := extensionsClient.CustomResourceDefinitions().Create(migrationsCRDV1Beta1())
			if err != nil {
				return errors.Wrap(err, "failed to create migrations crd")
			}
		}

		return nil
	}

	extensionsClient, err := extensionsv1client.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "faild to create extensions client")
	}

	_, err := extensionsClient.CustomResourceDefinitions().Get("migrations.schemas.schemahero.io", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get migrations crd")
		}

		_, err := extensionsClient.CustomResourceDefinitions().Create(migrationsCRDV1())
		if err != nil {
			return errors.Wrap(err, "failed to create migrations crd")
		}
	}

	return nil
}

func migrationsCRDV1Beta1() *extensionsv1beta1.CustomResourceDefinition {
	return nil
}

func migrationsCRDV1() *extensionsv1.CustomResourceDefinition {
	return nil
}
