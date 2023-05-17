
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

UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
define LDFLAGS
-ldflags "\
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
	-w -extldflags \"-static\" \
"
endef
else # all other OSes, including Windows and Darwin
define LDFLAGS
-ldflags "\
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef
endif

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
	GOOS=linux GOARCH=amd64 go build \
	  -tags netgo -installsuffix netgo \
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
	cp -R config/crds/v1/ pkg/installer/assets

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: generate
generate: controller-gen client-gen lister-gen informer-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/apis/...
	$(CLIENT_GEN) \
		--output-package=github.com/schemahero/schemahero/pkg/client \
		--clientset-name schemaheroclientset \
		--input-base github.com/schemahero/schemahero/pkg/apis \
		--input databases/v1alpha4 \
		--input schemas/v1alpha4 \
		-h ./hack/boilerplate.go.txt

	$(LISTER_GEN) \
        --input-dirs github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4,github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4 \
		--output-package=github.com/schemahero/schemahero/pkg/client/schemaherolisters \
        -h ./hack/boilerplate.go.txt

	$(INFORMER_GEN) \
		--output-package=github.com/schemahero/schemahero/pkg/client/schemaheroinformers \
		--listers-package=github.com/schemahero/schemahero/pkg/client/schemaherolisters \
		--versioned-clientset-package github.com/schemahero/schemahero/pkg/client/schemaheroclientset \
        --input-dirs github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4,github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4 \
        -h ./hack/boilerplate.go.txt

.PHONY: bin/kubectl-schemahero
bin/kubectl-schemahero:
	GOOS=linux GOARCH=amd64 go build \
	  -tags netgo -installsuffix netgo \
		${LDFLAGS} \
		-o bin/kubectl-schemahero \
		./cmd/kubectl-schemahero
	@echo "built bin/kubectl-schemahero"

.PHONY: local
local: bin/kubectl-schemahero manager
	docker build -t schemahero/schemahero-manager -f ./Dockerfile.multiarch --target manager .
	docker tag schemahero/schemahero-manager localhost:32000/schemahero/schemahero-manager:latest
	docker push localhost:32000/schemahero/schemahero-manager:latest

.PHONY: kind
kind: bin/kubectl-schemahero manager

.PHONY: contoller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: client-gen
client-gen:
ifeq (, $(shell which client-gen))
	go install k8s.io/code-generator/cmd/client-gen@kubernetes-1.25.3
CLIENT_GEN=$(shell go env GOPATH)/bin/client-gen
else
CLIENT_GEN=$(shell which client-gen)
endif

.PHONY: lister-gen
lister-gen:
ifeq (, $(shell which lister-gen))
	go install k8s.io/code-generator/cmd/lister-gen@kubernetes-1.25.3
LISTER_GEN=$(shell go env GOPATH)/bin/lister-gen
else
LISTER_GEN=$(shell which lister-gen)
endif

.PHONY: informer-gen
informer-gen:
ifeq (, $(shell which informer-gen))
	go install k8s.io/code-generator/cmd/informer-gen@kubernetes-1.25.3
INFORMER_GEN=$(shell go env GOPATH)/bin/informer-gen
else
INFORMER_GEN=$(shell which informer-gen)
endif

.PHONY: release-tarballs
release-tarballs:
	rm -rf release
	mkdir -p ./release

	# Build the kubectl plugins

	rm -rf ./bin/kubectl-schemahero
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_linux_amd64.tar.gz ./kubectl-schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_linux_arm64.tar.gz ./kubectl-schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero.exe
	tar czvf ./release/kubectl-schemahero_windows_amd64.tar.gz ./kubectl-schemahero.exe README.md LICENSE
	rm kubectl-schemahero.exe

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_darwin_amd64.tar.gz ./kubectl-schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_darwin_arm64.tar.gz ./kubectl-schemahero README.md LICENSE

	rm -rf ./kubectl-schemahero

.PHONY: build-manager
build-manager:
	CGO_ENABLED=0 make bin/manager

.PHONY: build-schemahero
build-schemahero:
	CGO_ENABLED=0 GOOS=linux make bin/kubectl-schemahero

.PHONY: cosign-sign
cosign-sign:
	# cosign attach sbom --sbom ./sbom/bom-go-mod.spdx schemahero/schemahero@${DIGEST_SCHEMAHERO}
	# cosign attach sbom --sbom ./sbom/bom-go-mod.spdx schemahero/schemahero-manager@${DIGEST_SCHEMAHERO_MANAGER}
	cosign sign --key ./cosign.key schemahero/schemahero@${DIGEST_SCHEMAHERO}
	cosign sign --key ./cosign.key schemahero/schemahero-manager@${DIGEST_SCHEMAHERO_MANAGER}

.PHONY: scan
scan:
	trivy fs \
		--security-checks vuln \
		--exit-code=1 \
		--severity="HIGH,CRITICAL" \
		--ignore-unfixed \
		./
