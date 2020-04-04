package installer

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GenerateOperatorYAML(requestedExtensionsAPIVersion string, isEnterprise bool, enterpriseTag string, namespace string) (map[string][]byte, error) {
	manifests := map[string][]byte{}

	useExtensionsV1Beta1 := false
	if requestedExtensionsAPIVersion == "v1beta1" {
		useExtensionsV1Beta1 = true
	}

	manifest, err := databasesCRDYAML(useExtensionsV1Beta1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases crd")
	}
	manifests["databases_crd.yaml"] = manifest

	manifest, err = tablesCRDYAML(useExtensionsV1Beta1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tables crd")
	}
	manifests["tables_crd.yaml"] = manifest

	manifest, err = migrationsCRDYAML(useExtensionsV1Beta1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get migrations crd")
	}
	manifests["migrations_crd.yaml"] = manifest

	manifest, err = clusterRoleYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster role")
	}
	manifests["cluster-role.yaml"] = manifest

	manifest, err = clusterRoleBindingYAML(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster binding role")
	}
	manifests["cluster-role-binding.yaml"] = manifest

	if !isEnterprise {
		manifest, err = namespaceYAML(namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get namespace")
		}
		manifests["namespace.yaml"] = manifest
	}

	manifest, err = serviceYAML(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service")
	}
	manifests["service.yaml"] = manifest

	manifest, err = secretYAML(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret")
	}
	manifests["secret.yaml"] = manifest

	manifest, err = managerYAML(isEnterprise, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manager")
	}
	manifests["manager.yaml"] = manifest

	if isEnterprise {
		manifest, err = heroAPIYAML(namespace, enterpriseTag)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get hero api")
		}
		manifests["hero-api.yaml"] = manifest

		manifest, err = heroAPIServiceYAML(namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get hero api service")
		}
		manifests["hero-api-service.yaml"] = manifest

		manifest, err = heroWebYAML(namespace, enterpriseTag)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get hero web")
		}
		manifests["hero-web.yaml"] = manifest

		manifest, err = heroWebServiceYAML(namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get hero web service")
		}
		manifests["hero-web-service.yaml"] = manifest
	}
	return manifests, nil
}

func InstallOperator(isEnterprise bool, namespace string) error {
	// todo create and pass this from higher
	ctx := context.Background()

	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes config")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create new kubernetes client")
	}

	useExtensionsV1Beta1 := shouldUseExtensionsV1Beta1(client)
	if err := ensureDatabasesCRD(ctx, cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create databases crd")
	}

	if err := ensureTablesCRD(ctx, cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create tables crd")
	}

	if err := ensureMigrationsCRD(ctx, cfg, useExtensionsV1Beta1); err != nil {
		return errors.Wrap(err, "failed to create migrations crd")
	}

	if err := ensureClusterRole(ctx, client); err != nil {
		return errors.Wrap(err, "failed to create cluster role")
	}

	if err := ensureClusterRoleBinding(ctx, client, namespace); err != nil {
		return errors.Wrap(err, "failed to create cluster role binding")
	}

	if err := ensureNamespace(ctx, client, namespace); err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	if err := ensureService(ctx, client, namespace); err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	if err := ensureSecret(ctx, client, namespace); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	if err := ensureManager(ctx, client, isEnterprise, namespace); err != nil {
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
