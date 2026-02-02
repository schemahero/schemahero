SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-postgres-plugin v24.3.25 v25.2.11 v25.4.3

.PHONY: build-postgres-plugin
build-postgres-plugin:
	@echo "Building postgres plugin (for cockroachdb)..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/postgres && go build -o $(HOME)/.schemahero/plugins/schemahero-postgres .

.PHONY: v24.3.25
v24.3.25: export COCKROACHDB_VERSION = v24.3.25
v24.3.25: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run

.PHONY: v25.2.11
v25.2.11: export COCKROACHDB_VERSION = v25.2.11
v25.2.11: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run

.PHONY: v25.4.3
v25.4.3: export COCKROACHDB_VERSION = v25.4.3
v25.4.3: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.multiarch --target --schemahero ..
	@echo "export IMAGE=$(IMAGE)"
