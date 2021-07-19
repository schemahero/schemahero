
SHELL := /bin/bash
VERSION ?=`git describe --tags`
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION_PACKAGE = github.com/schemahero/schemahero/pkg/version
GIT_TREE = $(shell git rev-parse --is-inside-work-tree 2>/dev/null)
ifneq "$(GIT_TREE)" ""
define GIT_UPDATE_INDEX_CMD
git update-index --assume-unchanged
endef
define GIT_SHA
`git rev-parse HEAD`
endef
else
define GIT_UPDATE_INDEX_CMD
echo "Not a git repo, skipping git update-index"
endef
define GIT_SHA
""
endef
endif

define LDFLAGS
-ldflags "\
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef

export GO111MODULE=on
# export GOPROXY=https://proxy.golang.org

all: generate fmt vet manifests bin/kubectl-schemahero manager test

.PHONY: clean-and-tidy
clean-and-tidy:
	@go clean -modcache ||:
	@go mod tidy ||:

.PHONY: envtest
envtest:
	./hack/envtest.sh

.PHONY: test
test: fmt vet manifests envtest
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: manager
manager: fmt vet bin/manager

.PHONY: bin/manager
bin/manager:
	go build \
		${LDFLAGS} \
		-o bin/manager \
		./cmd/manager

.PHONY: run
run: generate fmt vet bin/manager
	./bin/manager run \
	--log-level debug \
	--database-name="*"

.PHONY: run-database
run-database: generate fmt vet bin/manager
	./bin/manager run \
	--enable-database-controller \
	--manager-image localhost:32000/schemahero/schemahero-manager \
	--manager-tag latest

.PHONY: install
install: manifests generate local
	kubectl apply -f config/crds/v1

.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crds/v1
	kustomize build config/default | kubectl apply -f -

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) \
		rbac:roleName=manager-role webhook \
		crd:crdVersions=v1,generateEmbeddedObjectMeta=true  \
		output:crd:artifacts:config=config/crds/v1 \
		paths="./..."
	go run ./generate/...

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: generate
generate: controller-gen client-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/apis/...
	$(CLIENT_GEN) \
		--output-package=github.com/schemahero/schemahero/pkg/client \
		--clientset-name schemaheroclientset \
		--input-base github.com/schemahero/schemahero/pkg/apis \
		--input databases/v1alpha4 \
		--input schemas/v1alpha4 \
		-h ./hack/boilerplate.go.txt

.PHONY: bin/kubectl-schemahero
bin/kubectl-schemahero:
	go build \
		${LDFLAGS} \
		-o bin/kubectl-schemahero \
		./cmd/kubectl-schemahero
	@echo "built bin/kubectl-schemahero"

.PHONY: local
local: bin/kubectl-schemahero manager
	docker build -t schemahero/schemahero-manager -f ./Dockerfile.manager .
	docker tag schemahero/schemahero-manager localhost:32000/schemahero/schemahero-manager:latest
	docker push localhost:32000/schemahero/schemahero-manager:latest

.PHONY: kind
kind: bin/kubectl-schemahero manager

.PHONY: contoller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.0
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: client-gen
client-gen:
ifeq (, $(shell which client-gen))
	go install k8s.io/code-generator/cmd/client-gen@kubernetes-1.20.0
CLIENT_GEN=$(shell go env GOPATH)/bin/client-gen
else
CLIENT_GEN=$(shell which client-gen)
endif

.PHONY: sbom
sbom: spdx-generator
	mkdir -p sbom
	$(SPDX_GENERATOR) -o ./sbom

.PHONY: spdx-generator
spdx-generator:
ifeq (, $(shell which spdx-sbom-generator))
	mkdir -p sbom
	curl -L https://github.com/spdx/spdx-sbom-generator/releases/download/v0.0.10/spdx-sbom-generator-v0.0.10-linux-amd64.tar.gz -o ./sbom/spdx-sbom-generator-v0.0.10-linux-amd64.tar.gz
	tar xzvf ./sbom/spdx-sbom-generator-v0.0.10-linux-amd64.tar.gz -C sbom
SPDX_GENERATOR=./sbom/spdx-sbom-generator
else
SPDX_GENERATOR=$(shell which spdx-sbom-generator)
endif

.PHONY: release
release:
	rm -rf release
	mkdir -p ./release

	# Build the kubectl plugins

	rm -rf ./bin/kubectl-schemahero
	GOOS=linux GOARCH=amd64 make bin/kubectl-schemahero
	tar czvf ./release/kubectl-schemahero-linux-amd64.tar.gz ./bin/kubectl-schemahero README.md LICENSE

	rm -rf ./bin/kubectl-schemahero
	GOOS=linux GOARCH=arm64 make bin/kubectl-schemahero
	tar czvf ./release/kubectl-schemahero-linux-arm64.tar.gz ./bin/kubectl-schemahero README.md LICENSE

	rm -rf ./bin/kubectl-schemahero
	GOOS=windows GOARCH=amd64 make bin/kubectl-schemahero
	tar czvf ./release/kubectl-schemahero-windows-amd64.tar.gz ./bin/kubectl-schemahero README.md LICENSE

	rm -rf ./bin/kubectl-schemahero
	GOOS=darwin GOARCH=amd64 make bin/kubectl-schemahero
	tar czvf ./release/kubectl-schemahero-darwin-amd64.tar.gz ./bin/kubectl-schemahero README.md LICENSE

	rm -rf ./bin/kubectl-schemahero
	GOOS=darwin GOARCH=arm64 make bin/kubectl-schemahero
	tar czvf ./release/kubectl-schemahero-darwin-arm64.tar.gz ./bin/kubectl-schemahero README.md LICENSE

	# build the docker images for in-cluster

	GOOS=linux GOARCH=amd64 make bin/manager
	GOOS=linux GOARCH=amd64 make bin/kubectl-schemahero
	docker build -t schemahero/schemahero:${GITHUB_TAG} -f ./deploy/Dockerfile.schemahero .
	docker build -t schemahero/manager:${GITHUB_TAG} -f ./deploy/Dockerfile.manager .
	docker push schemahero/schemahero:${GITHUB_TAG}
	docker push schemahero/manager:${GITHUB_TAG}
	cosign attach sbom -sbom ./sbom/bom-go-mod.spdx schemahero/schemahero:${GITHUB_TAG}
	cosign attach sbom -sbom ./sbom/bom-go-mod.spdx schemahero/manager:${GITHUB_TAG}
	cosign sign -key ./cosign.key schemahero/schemahero:${GITHUB_TAG}
	cosign sign -key ./cosign.key schemahero/manager:${GITHUB_TAG}
