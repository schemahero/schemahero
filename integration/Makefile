export GO_BUILD=env GO111MODULE=on go build
MANAGER_IMAGE_NAME := $(shell uuidgen)
SCHEMAHERO_IMAGE_NAME := $(shell uuidgen)

.PHONY: run
run: build integration-test-image
	bin/schemahero-integration-tests run \
		--manager-image-name "ttl.sh/$(MANAGER_IMAGE_NAME):1h" \
		--schemahero-image-name "ttl.sh/$(SCHEMAHERO_IMAGE_NAME):1h"

integration-test-image:
	docker build -t ttl.sh/$(SCHEMAHERO_IMAGE_NAME):1h -f ../Dockerfile.schemahero ..
	docker build -t ttl.sh/$(MANAGER_IMAGE_NAME):1h -f ../Dockerfile.manager ..
	docker push ttl.sh/$(SCHEMAHERO_IMAGE_NAME):1h
	docker push ttl.sh/$(MANAGER_IMAGE_NAME):1h

.PHONY: build
build: GO111MODULE = "on"
build:
	rm -rf bin/schemahero-integration-tests
	$(GO_BUILD) \
		-ldflags "\
			-X ${VERSION_PACKAGE}.version=${VERSION} \
			-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
			-X ${VERSION_PACKAGE}.buildTime=${DATE}" \
		-o bin/schemahero-integration-tests \
		./cmd/integration
	@echo "built bin/schemahero-integration-tests"
