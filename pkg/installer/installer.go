package installer

import (
	"github.com/pkg/errors"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func InstallOperator() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes config")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create new kubernetes client")
	}

	extensionsClient, err := extensionsclient.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "faield to create extensions client")
	}

	if err := createDatabasesCRD(client, extensionsClient); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	return nil
}

func createDatabasesCRD(client *kubernetes.Clientset, extensionsClient *extensionsclient.ApiextensionsV1Client) error {
	databasesCRD := extensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "databases.databases.schemahero.io",
		},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Group: "databases.schemahero.io",
		},
	}

	if _, err := extensionsClient.CustomResourceDefinitions().Create(&databasesCRD); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	return nil
}
