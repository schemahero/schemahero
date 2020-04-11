
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: postgres mysql cockroach

.PHONY: postgres
postgres:
	make -C tests/postgres/column-set-default run
	make -C tests/postgres/column-unset-default run
	make -C tests/postgres/create-table run
	make -C tests/postgres/foreign-key-create run
	make -C tests/postgres/foreign-key-action run
	make -C tests/postgres/foreign-key-drop run
	make -C tests/postgres/foreign-key-alter run
	make -C tests/postgres/not-null run
	make -C tests/postgres/not-null-with-default run
	make -C tests/postgres/index-create run
	make -C tests/postgres/primary-key-add run
	make -C tests/postgres/primary-key-drop run
	make -C tests/postgres/unique-constraint-add run
	make -C tests/postgres/unique-constraint-drop run

.PHONY: mysql
mysql:
	make -C tests/mysql/column-set-default run
	make -C tests/mysql/column-unset-default run
	make -C tests/mysql/create-table run
	make -C tests/mysql/foreign-key-create run
	make -C tests/mysql/foreign-key-action run
	make -C tests/mysql/foreign-key-drop run
	make -C tests/mysql/foreign-key-alter run
	make -C tests/mysql/not-null run
	make -C tests/mysql/index-create run
	make -C tests/mysql/primary-key-add run
	make -C tests/mysql/primary-key-drop run

.PHONY: cockroach
cockroach:
	make -C tests/cockroach run TEST_NAME=column-set-default
	make -C tests/cockroach run TEST_NAME=column-unset-default
	make -C tests/cockroach run TEST_NAME=create-table
	make -C tests/cockroach run TEST_NAME=foreign-key-create
	make -C tests/cockroach run TEST_NAME=foreign-key-action
	make -C tests/cockroach run TEST_NAME=foreign-key-drop
	make -C tests/cockroach run TEST_NAME=foreign-key-alter
	make -C tests/cockroach run TEST_NAME=not-null
	make -C tests/cockroach run TEST_NAME=index-create
	make -C tests/cockroach run TEST_NAME=primary-key-add
	make -C tests/cockroach run TEST_NAME=primary-key-drop

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
