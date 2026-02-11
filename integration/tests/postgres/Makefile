SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-postgres-plugin 14.20 15.15 16.11 17.7 18.1

.PHONY: build-postgres-plugin
build-postgres-plugin:
	@echo "Building postgres plugin..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/postgres && go build -o $(HOME)/.schemahero/plugins/schemahero-postgres .

.PHONY: 14.20
14.20: export PG_VERSION = 14.20
14.20: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C create-table-with-index run
	make -C drop-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C foreign-key-idempotent run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C two-column-pk run
	make -C two-column-pk-reversed-order run
	make -C user-defined-type-vector run
	make -C batch-tables run

.PHONY: 15.15
15.15: export PG_VERSION = 15.15
15.15: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-function run
	make -C create-table run
	make -C create-table-with-index run
	make -C drop-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C foreign-key-idempotent run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C two-column-pk run
	make -C two-column-pk-reversed-order run
	make -C user-defined-type-vector run
	make -C batch-tables run

.PHONY: 16.11
16.11: export PG_VERSION = 16.11
16.11: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-function run
	make -C create-table run
	make -C create-table-with-index run
	make -C drop-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C foreign-key-idempotent run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C two-column-pk run
	make -C two-column-pk-reversed-order run
	make -C user-defined-type-vector run
	make -C batch-tables run

.PHONY: 17.7
17.7: export PG_VERSION = 17.7
17.7: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-function run
	make -C create-table run
	make -C create-table-with-index run
	make -C drop-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C foreign-key-idempotent run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C two-column-pk run
	make -C two-column-pk-reversed-order run
	make -C user-defined-type-vector run
	make -C batch-tables run

.PHONY: 18.1
18.1: export PG_VERSION = 18.1
18.1: build-postgres-plugin
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-function run
	make -C create-table run
	make -C create-table-with-index run
	make -C drop-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C foreign-key-idempotent run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C two-column-pk run
	make -C two-column-pk-reversed-order run
	make -C user-defined-type-vector run
	make -C batch-tables run

.PHONY: seed
seed: export PG_VERSION = 17.7
seed: build-postgres-plugin
	make -C basic-seed run
	make -C seed-data-with-schema run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.multiarch --target --schemahero ..
	@echo "export IMAGE=$(IMAGE)"
