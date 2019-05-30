
SHELL := /bin/bash
VERSION ?=`git describe --tags`
FULLSRC = $(shell find pkg vendor -name "*.go")
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

all: test bin/schemahero manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet bin/manager

bin/manager: $(FULLSRC) cmd/manager/main.go
	go build \
		${LDFLAGS} \
		-i \
		-o bin/manager \
		./cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet bin/schemahero
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests microk8s
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: generate
generate:
ifndef GOPATH
	$(error GOPATH not defined, please define GOPATH. Run "go help gopath" to learn more about GOPATH)
endif
	go generate ./pkg/... ./cmd/...
	rm -r ./pkg/client/schemaheroclientset/fake

.PHONY: integration/postgres
integration/postgres: bin/schemahero
	@-docker rm -f schemahero-postgres > /dev/null 2>&1 ||:
	docker pull postgres:10
	docker run --rm -d --name schemahero-postgres -p 15432:5432 \
		-e POSTGRES_PASSWORD=password \
		-e POSTGRES_USER=schemahero \
		-e POSTGRES_DB=schemahero \
		postgres:10
	@-sleep 5
	./bin/schemahero watch --driver postgres --uri postgres://schemahero:password@localhost:15432/schemahero?sslmode=disable
	docker rm -f schemahero-postgres

bin/schemahero: $(FULLSRC) cmd/schemahero/main.go
	go build \
		${LDFLAGS} \
		-i \
		-o bin/schemahero \
		./cmd/schemahero
	@echo "built bin/schemahero"

.PHONY: docker-login
docker-login:
	echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

.PHONY: installable-manifests-snapshot
installable-manifests-snapshot:
	cd config/default; kustomize edit set image schemahero/schemahero-manager:alpha
	kustomize build config/default > install/schemahero/schemahero-operator.yaml

.PHONY: snapshot-release
snapshot-release: build-snapshot-release installable-manifests-snapshot
	docker push schemahero/schemahero:alpha
	docker push schemahero/schemahero-manager:alpha
	@echo "Manifests were updated in this repo. Push to make sure they are live."

.PHONY: build-snapshot-release
build-snapshot-release:
	curl -sL https://git.io/goreleaser | bash -s -- --rm-dist --snapshot --config deploy/.goreleaser.snapshot.yml

.PHONY: microk8s
microk8s:
	docker build -t schemahero/schemahero -f ./Dockerfile.schemahero .
	docker tag schemahero/schemahero localhost:32000/schemahero/schemahero:latest
	docker push localhost:32000/schemahero/schemahero:latest
