
SHELL := /bin/bash
CONSTIMAGE := "ttl.sh/$(shell uuidgen | tr '[:upper:]' '[:lower:]'):1h"
IMAGE ?= $(CONSTIMAGE)

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build postgres mysql

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

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
