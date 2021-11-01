package installer

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	extensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
)

//go:embed assets/schemas.schemahero.io_migrations.yaml
var generatedMigrationCRDV1 string

func migrationsCRDYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer

	if err := s.Encode(migrationsCRDV1(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal migrations v1 crd")
	}

	return result.Bytes(), nil
}

func ensureMigrationsCRD(ctx context.Context, cfg *rest.Config) error {
	extensionsClient, err := extensionsv1client.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "faild to create extensions client")
	}

	existingCRD, err := extensionsClient.CustomResourceDefinitions().Get(ctx, "migrations.schemas.schemahero.io", metav1.GetOptions{})

	// if there's an error and it's not a NotFound error, that's unexpected and we cannot continue
	if err != nil && !kuberneteserrors.IsNotFound(err) {
		return errors.Wrap(err, "get migrations crd")
	}

	if kuberneteserrors.IsNotFound(err) {
		_, err := extensionsClient.CustomResourceDefinitions().Create(ctx, migrationsCRDV1(), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create migrations crd")
		}
		return nil
	}

	// update the existing object with the new
	existingCRD.Spec = migrationsCRDV1().Spec
	existingCRD.Labels = migrationsCRDV1().Labels
	existingCRD.Annotations = migrationsCRDV1().Annotations

	_, err = extensionsClient.CustomResourceDefinitions().Update(ctx, existingCRD, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "update migrations crd")
	}

	return nil
}

func migrationsCRDV1() *extensionsv1.CustomResourceDefinition {
	extensionsscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(generatedMigrationCRDV1), nil, nil)
	if err != nil {
		panic(err) // todo
	}

	return obj.(*extensionsv1.CustomResourceDefinition)
}
