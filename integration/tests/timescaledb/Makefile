
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 2.8.1-pg14

.PHONY: 2.8.1-pg14
2.8.1-pg14: export TIMESCALEDB_VERSION = 2.8.1-pg14
2.8.1-pg14:
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


.PHONY: seed
seed: export TIMESCALEDB_VERSION = 2.8.1-pg14
seed:
	make -C basic-seed run
	make -C seed-with-many-rows run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
