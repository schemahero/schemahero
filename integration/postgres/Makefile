SHELL := /bin/bash
CLUSTER_NAME := schemahero-integration-postgres
MANAGER_IMAGE_NAME := schemahero-integration/manager:postgres

.PHONY: test
test: cleanup-before cluster crd cleanup-after

.PHONY: cluster
cluster:
	kind create cluster --wait 90s --name $(CLUSTER_NAME)
	KUBECONFIG=$$(kind get kubeconfig-path --name="$(CLUSTER_NAME)") kubectl get nodes

.PHONY: cleanup-%
cleanup-%:
	-kind delete cluster --name $(CLUSTER_NAME) > /dev/null 2>&1 ||:

.PHONY: crd
crd:
	@echo "building manager image"
	docker build -t $(MANAGER_IMAGE_NAME) -f ../../Dockerfile ../../

	@echo "updating kustomize image patch to use newly built image"
	sed -i'' -e 's@image: .*@image: '"${MANAGER_IMAGE_NAME}"'@' ../../config/integration/manager_image_patch.yaml

	KUBECONFIG=$$(kind get kubeconfig-path --name="$(CLUSTER_NAME)") kubectl apply -f ../../config/crds
