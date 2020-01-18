package installer

import (
	"bytes"

	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GenerateOperatorYAML() ([]byte, error) {
	manifests := [][]byte{}

	manifest, err := databasesCRDYAML(false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases crd")
	}
	manifests = append(manifests, manifest)

	manifest, err = tablesCRDYAML(false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tables crd")
	}
	manifests = append(manifests, manifest)

	manifest, err = migrationsCRDYAML(false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get migrations crd")
	}
	manifests = append(manifests, manifest)

	manifest, err = clusterRoleYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster role")
	}
	manifests = append(manifests, manifest)

	manifest, err = clusterRoleBindingYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster binding role")
	}
	manifests = append(manifests, manifest)

	manifest, err = namespaceYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namespace")
	}
	manifests = append(manifests, manifest)

	manifest, err = serviceYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service")
	}
	manifests = append(manifests, manifest)

	manifest, err = secretYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret")
	}
	manifests = append(manifests, manifest)

	manifest, err = managerYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manager")
	}
	manifests = append(manifests, manifest)

	multiDocResult := bytes.Join(manifests, []byte("\n---\n"))

	return multiDocResult, nil
}

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
		return errors.Wrap(err, "faild to create extensions client")
	}

	useExtensionsV1Beta1 := shouldUseExtensionsV1Beta1(client)
	if err := ensureDatabasesCRD(cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	if err := ensureTablesCRD(cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create tables crd")
	}

	if err := ensureMigrationsCRD(cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create migrations crd")
	}

	if err := ensureClusterRole(client); err != nil {
		return errors.Wrap(err, "failed to create cluster role")
	}

	if err := ensureClusterRoleBinding(client); err != nil {
		return errors.Wrap(err, "failed to create cluster role binding")
	}

	if err := ensureNamespace(client); err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	if err := ensureService(client); err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	if err := ensureSecret(client); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	if err := ensureManager(client); err != nil {
		return errors.Wrap(err, "failed to create manager")
	}

	return nil
}

func shouldUseExtensionsV1Beta1(client *kubernetes.Clientset) bool {
	// if there's no client or no server, just return v1, it's not an error
	serverVersion, err := client.ServerVersion()
	if err != nil {
		return false
	}

	parsedVersion, err := semver.Make(strings.TrimLeft(serverVersion.String(), "v"))
	if err != nil {
		return false
	}
	minimumExtensionsV1 := semver.MustParse("1.16.0")
	return parsedVersion.LT(minimumExtensionsV1)
}
