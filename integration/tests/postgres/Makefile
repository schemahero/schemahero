
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 10.7 11.8 12.3

.PHONY: 10.13
10.13: export PG_VERSION = 10.13
10.13:
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

.PHONY: 11.8
11.8: export PG_VERSION = 11.8
11.8:
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

.PHONY: 12.3
12.3: export PG_VERSION = 12.3
12.3:
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

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
