
SHELL := /bin/bash
VERSION ?= alpha

all: test bin/schemahero manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/schemahero/schemahero/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet bin/schemahero
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
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

.PHONY: bin/schemahero
bin/schemahero:
	rm -rf bin/schemahero
	go build \
		-ldflags "\
			-X ${VERSION_PACKAGE}.version=${VERSION} \
			-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
			-X ${VERSION_PACKAGE}.buildTime=${DATE}" \
		-o bin/schemahero \
		./cmd/schemahero
	@echo "built bin/schemahero"

.PHONY: docker-login:
docker-login:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin

.PHONY: installable-manifests
installable-manifests:
	cd config/default; kustomize edit set image schemahero/schemahero-manager:${VERSION}
	kustomize build config/default > install/schemahero/schemahero-operator.yaml
	cd config/default; git checkout .

.PHONY: snapshot-release
snapshot-release: build-snapshot-release installable-manifests
	docker push schemahero/schemahero:alpha
	docker push schemahero/schemahero-manager:alpha
	@echo "Manifests were updated in this repo. Push to make sure they are live."

.PHONY: build-snapshot-release
build-snapshot-release:
	curl -sL https://git.io/goreleaser | bash -s -- --rm-dist --snapshot --config deploy/.goreleaser.snapshot.yml

.PHONY: micok8s
microk8s: build-release
	docker tag schemahero/schemahero localhost:32000/schemahero/schemahero:latest
	docker push localhost:32000/schemahero/schemahero:latest
