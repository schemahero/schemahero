package installer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/config"
	"k8s.io/client-go/kubernetes"
)

func GenerateOperatorYAML(namespace string) (map[string][]byte, error) {
	manifests := map[string][]byte{}

	manifest, err := databasesCRDYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases crd")
	}
	manifests["databases_crd.yaml"] = manifest

	manifest, err = tablesCRDYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tables crd")
	}
	manifests["tables_crd.yaml"] = manifest

	manifest, err = viewsCRDYAML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get views crd")
	}
	manifests["views_crd.yaml"] = manifest

	manifest, err = migrationsCRDYAML()
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

	manifest, err = namespaceYAML(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namespace")
	}
	manifests["namespace.yaml"] = manifest

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

	manifest, err = managerYAML(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manager")
	}
	manifests["manager.yaml"] = manifest

	return manifests, nil
}

func InstallOperator(namespace string) (bool, error) {
	// todo create and pass this from higher
	ctx := context.Background()

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return false, errors.Wrap(err, "failed to get kubernetes config")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, errors.Wrap(err, "failed to create new kubernetes client")
	}

	if err := ensureDatabasesCRD(ctx, cfg); err != nil {
		return false, errors.Wrap(err, "failed to create databases crd")
	}

	if err := ensureTablesCRD(ctx, cfg); err != nil {
		return false, errors.Wrap(err, "failed to create tables crd")
	}

	if err := ensureViewsCRD(ctx, cfg); err != nil {
		return false, errors.Wrap(err, "failed to create views crd")
	}

	if err := ensureMigrationsCRD(ctx, cfg); err != nil {
		return false, errors.Wrap(err, "failed to create migrations crd")
	}

	if err := ensureClusterRole(ctx, client); err != nil {
		return false, errors.Wrap(err, "failed to create cluster role")
	}

	if err := ensureClusterRoleBinding(ctx, client, namespace); err != nil {
		return false, errors.Wrap(err, "failed to create cluster role binding")
	}

	if err := ensureNamespace(ctx, client, namespace); err != nil {
		return false, errors.Wrap(err, "failed to create namespace")
	}

	if err := ensureService(ctx, client, namespace); err != nil {
		return false, errors.Wrap(err, "failed to create service")
	}

	if err := ensureSecret(ctx, client, namespace); err != nil {
		return false, errors.Wrap(err, "failed to create secret")
	}

	wasUpgraded, err := ensureManager(ctx, client, namespace)
	if err != nil {
		return false, errors.Wrap(err, "failed to create manager")
	}

	return wasUpgraded, nil
}
