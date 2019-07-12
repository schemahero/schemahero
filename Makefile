
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

all: test bin/schemahero manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet bin/manager

bin/manager:
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

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) paths=./pkg/apis/... output:dir=./config/crds

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/api/...

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

bin/schemahero:
	go build \
		${LDFLAGS} \
		-i \
		-o bin/schemahero \
		./cmd/schemahero
	@echo "built bin/schemahero"

.PHONY: docker-login
docker-login:
	echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

.PHONY: snapshot-release
snapshot-release: build-snapshot-release
	docker push schemahero/schemahero:alpha
	docker push schemahero/schemahero-manager:alpha

.PHONY: build-snapshot-release
build-snapshot-release:
	curl -sL https://git.io/goreleaser | bash -s -- --rm-dist --snapshot --config deploy/.goreleaser.snapshot.yml

.PHONY: microk8s
microk8s:
	docker build -t schemahero/schemahero -f ./Dockerfile.schemahero .
	docker tag schemahero/schemahero localhost:32000/schemahero/schemahero:latest
	docker push localhost:32000/schemahero/schemahero:latest

.PHONY: tag-release
tag-release:
	curl -sL https://git.io/goreleaser | bash -s -- --rm-dist --config deploy/.goreleaser.yml

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0-beta.2
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
