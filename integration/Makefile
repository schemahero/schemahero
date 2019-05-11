export GO_BUILD=env GO111MODULE=on go build
MANAGER_IMAGE_NAME := $(shell uuidgen)
SCHEMA_HERO_IMAGE_NAME := $(shell uuidgen)

.PHONY: run
run: build integration-test-image
	bin/schemahero-integration-tests run

integration-test-image:
	cd .. && curl -sL https://git.io/goreleaser | bash -s -- --snapshot --rm-dist --config deploy/.goreleaser.integration.yml
	docker tag schemahero/schemahero:integration-test rgstry.dev/$(SCHEMA_HERO_IMAGE_NAME):1h
	docker tag schemahero/schemahero-manager:integration-test rgstry.dev/$(MANAGER_IMAGE_NAME):1h
	docker push rgstry.dev/$(SCHEMA_HERO_IMAGE_NAME):1h
	docker push rgstry.dev/$(MANAGER_IMAGE_NAME):1h

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
