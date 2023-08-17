
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 11.18 12.13 13.9 14.6 15.1

.PHONY: 11.18
11.18: export PG_VERSION = 11.18-alpine
11.18:
	make -C alter-column-timezone run
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: 12.13
12.13: export PG_VERSION = 12.13
12.13:
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: 13.9
13.9: export PG_VERSION = 13.9
13.9:
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: 14.6
14.6: export PG_VERSION = 14.6
14.6:
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: 15.1
15.1: export PG_VERSION = 15.1
15.1:
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C not-null-with-default run
	make -C index-create run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C unique-constraint-add run
	make -C unique-constraint-drop run
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: seed
seed: export PG_VERSION = 15.1
seed:
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.multiarch --target --schemahero ..
	@echo "export IMAGE=$(IMAGE)"
