
SHELL := /bin/bash
VERSION ?= $(if $(GIT_TAG),$(GIT_TAG),$(shell git describe --tags))
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION_PACKAGE = github.com/schemahero/schemahero/pkg/version
PLUGIN ?= postgres
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

.PHONY: help
help: ## Show this help message
	@echo "SchemaHero Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: clean-and-tidy
clean-and-tidy: ## Clean module cache and tidy dependencies
	@go clean -modcache ||:
	@go mod tidy ||:

.PHONY: envtest
envtest:
	./hack/envtest.sh

.PHONY: test
test: fmt vet manifests envtest ## Run tests
	go test ./pkg/... ./cmd/...

.PHONY: plugins
plugins: ## Build all database plugins
	@echo "Building plugins..."
	$(MAKE) -C plugins all

.PHONY: install-dev
install-dev: bin/kubectl-schemahero
	@echo "Building $(PLUGIN) plugin for development..."
	$(MAKE) -C plugins $(PLUGIN)
	@echo ""
	@echo "✅ Development build complete!"
	@echo "🔧 Binary: ./bin/kubectl-schemahero"
	@echo "🔌 Plugin: ./plugins/bin/schemahero-$(PLUGIN)"
	@echo ""
	@echo "To use:"
	@echo "  SCHEMAHERO_PLUGIN_DIR=./plugins/bin ./bin/kubectl-schemahero plan ..."

.PHONY: install-plugin
install-plugin:
	@echo "Installing $(PLUGIN) plugin for development..."
	$(MAKE) -C plugins $(PLUGIN)
	sudo mkdir -p /var/lib/schemahero/plugins
	sudo cp ./plugins/bin/schemahero-$(PLUGIN) /var/lib/schemahero/plugins
	@echo "Fixing permissions and signing in place..."
	sudo chmod 755 /var/lib/schemahero/plugins/schemahero-$(PLUGIN)
	sudo codesign --force --deep -s - /var/lib/schemahero/plugins/schemahero-$(PLUGIN)
	@echo "Plugin $(PLUGIN) installed at /var/lib/schemahero/plugins/schemahero-$(PLUGIN)"

.PHONY: test-plugins
test-plugins:
	@echo "Testing plugins..."
	$(MAKE) -C plugins test

.PHONY: manager
manager: fmt vet bin/manager

.PHONY: bin/manager
bin/manager:
	go build \
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
install: manifests generate local ## Install SchemaHero to cluster
	kubectl apply -f config/crds/v1

.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crds/v1
	kustomize build config/default | kubectl apply -f -

.PHONY: manifests
manifests: controller-gen ## Generate Kubernetes manifests
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

.PHONY: tidy
tidy: go.mod plugins/**/go.mod
	for file in $^ ; do \
		dir=$$(dirname $$file); \
		pushd $$(dirname $$file) && go mod tidy; popd; \
	done

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
bin/kubectl-schemahero: ## Build kubectl-schemahero binary
	go build \
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

.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0
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
	mv ./kubectl-schemahero ./schemahero
	tar czvf ./release/schemahero_linux_amd64.tar.gz ./schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_linux_arm64.tar.gz ./kubectl-schemahero README.md LICENSE
	mv ./kubectl-schemahero ./schemahero
	tar czvf ./release/schemahero_linux_arm64.tar.gz ./schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero.exe
	tar czvf ./release/kubectl-schemahero_windows_amd64.tar.gz ./kubectl-schemahero.exe README.md LICENSE
	mv ./kubectl-schemahero.exe ./schemahero.exe
	tar czvf ./release/schemahero_windows_amd64.tar.gz ./schemahero.exe README.md LICENSE
	rm -rf ./kubectl-schemahero.exe
	rm -rf ./schemahero.exe

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_darwin_amd64.tar.gz ./kubectl-schemahero README.md LICENSE
	mv ./kubectl-schemahero ./schemahero
	tar czvf ./release/schemahero_darwin_amd64.tar.gz ./schemahero README.md LICENSE

	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 make bin/kubectl-schemahero
	mv bin/kubectl-schemahero ./kubectl-schemahero
	tar czvf ./release/kubectl-schemahero_darwin_arm64.tar.gz ./kubectl-schemahero README.md LICENSE
	mv ./kubectl-schemahero ./schemahero
	tar czvf ./release/schemahero_darwin_arm64.tar.gz ./schemahero README.md LICENSE

	rm -rf ./kubectl-schemahero
	rm -rf ./schemahero

.PHONY: build-manager
build-manager:
	CGO_ENABLED=0 make bin/manager

.PHONY: build-schemahero
build-schemahero:
	CGO_ENABLED=0 GOOS=linux make bin/kubectl-schemahero

.PHONY: cosign-sign
cosign-sign:
	# cosign attach sbom --sbom ./sbom/bom-go-mod.spdx ${DIGEST_SCHEMAHERO}
	# cosign attach sbom --sbom ./sbom/bom-go-mod.spdx ${DIGEST_SCHEMAHERO_MANAGER}
	cosign sign --yes --key ./cosign.key ${DIGEST_SCHEMAHERO}
	cosign sign --yes --key ./cosign.key ${DIGEST_SCHEMAHERO_MANAGER}

.PHONY: scan
scan:
	trivy fs \
		--security-checks vuln \
		--exit-code=1 \
		--severity="HIGH,CRITICAL" \
		--ignore-unfixed \
		./

.PHONY: test-dev
test-dev: ## Build SchemaHero for local testing with temporary ttl.sh images
	@echo "Building SchemaHero for local testing with ttl.sh images..."
	
	# Generate deterministic image name based on MAC address for ttl.sh (24 hour TTL)
	$(eval MAC_ADDR := $(shell ifconfig | grep ether | head -1 | awk '{print $$2}' | tr -d ':' | head -c 12))
	$(eval TTL_IMAGE := ttl.sh/schemahero-dev-$(MAC_ADDR))
	$(eval TTL_PLUGIN_PREFIX := ttl.sh/schemahero-dev-$(MAC_ADDR)/plugin)
	$(eval TEST_VERSION := dev)
	
	@echo "Using deterministic namespace: ttl.sh/schemahero-dev-$(MAC_ADDR)"
	@echo "Manager image: $(TTL_IMAGE):$(TEST_VERSION)"
	@echo "Plugin prefix: $(TTL_PLUGIN_PREFIX)-{driver}:$(TEST_VERSION)"
	
	# Build manager binary
	CGO_ENABLED=0 make bin/manager
	
	# Build multi-arch docker image and push to ttl.sh
	@echo "Setting up docker buildx for multi-platform builds..."
	docker buildx create --name schemahero-builder --use --bootstrap 2>/dev/null || docker buildx use schemahero-builder
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(TTL_IMAGE):$(TEST_VERSION) \
		-f ./deploy/Dockerfile.multiarch --target manager \
		--push .
	
	# Build and push multi-arch plugins to ttl.sh in parallel
	@echo "Building and pushing multi-arch plugins to ttl.sh in parallel..."
	@for plugin in postgres mysql timescaledb sqlite rqlite cassandra; do \
		if [ -d "./plugins/$$plugin" ]; then \
			echo "Building multi-arch plugin: $$plugin"; \
			( \
				echo "Building $$plugin for linux/amd64..." && \
				mkdir -p /tmp/push-$$plugin-amd64 && \
				cd ./plugins/$$plugin && \
				CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/push-$$plugin-amd64/schemahero-$$plugin . && \
				cd /tmp/push-$$plugin-amd64 && \
				oras push $(TTL_PLUGIN_PREFIX)-$$plugin:$(TEST_VERSION)-amd64 schemahero-$$plugin && \
				echo "Building $$plugin for linux/arm64..." && \
				mkdir -p /tmp/push-$$plugin-arm64 && \
				cd $(pwd)/plugins/$$plugin && \
				CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /tmp/push-$$plugin-arm64/schemahero-$$plugin . && \
				cd /tmp/push-$$plugin-arm64 && \
				oras push $(TTL_PLUGIN_PREFIX)-$$plugin:$(TEST_VERSION)-arm64 schemahero-$$plugin && \
				echo "✅ Pushed $$plugin for amd64 and arm64" \
			) & \
		else \
			echo "Warning: Plugin $$plugin directory not found, skipping"; \
		fi \
	done; \
	wait
	
	# Build kubectl-schemahero with embedded ttl.sh image and plugin registry
	go build \
		-tags netgo -installsuffix netgo \
		-ldflags "\
			-X github.com/schemahero/schemahero/pkg/version.version=$(TEST_VERSION) \
			-X github.com/schemahero/schemahero/pkg/version.gitSHA=`git rev-parse HEAD` \
			-X github.com/schemahero/schemahero/pkg/version.buildTime=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
			-X github.com/schemahero/schemahero/pkg/version.managerImage=$(TTL_IMAGE) \
			-X github.com/schemahero/schemahero/pkg/version.pluginRegistry=$(TTL_PLUGIN_PREFIX) \
		" \
		-o bin/kubectl-schemahero-test \
		./cmd/kubectl-schemahero
	
	@echo ""
	@echo "✅ Test build complete!"
	@echo "📦 Manager image: $(TTL_IMAGE):$(TEST_VERSION)"
	@echo "🔌 Plugin registry: $(TTL_PLUGIN_PREFIX)-{driver}:$(TEST_VERSION)"
	@echo "🔧 Binary: bin/kubectl-schemahero-test"
	@echo ""
	@echo "Plugins pushed:"
	@for plugin in postgres mysql timescaledb sqlite rqlite cassandra; do \
		if [ -f "./plugins/bin/schemahero-$$plugin" ]; then \
			echo "  • $(TTL_PLUGIN_PREFIX)-$$plugin:$(TEST_VERSION)"; \
		fi \
	done
	@echo ""
	@echo "To test:"
	@echo "  ./bin/kubectl-schemahero-test install"
	@echo ""
	@echo "🕐 Images will be available for 24 hours on ttl.sh"
