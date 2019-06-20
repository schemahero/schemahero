
SHELL := /bin/bash
IMAGE := "ttl.sh/$(shell uuidgen):1h"
export IMAGE
export GO111MODULE=on

.PHONY: run
run: build postgres mysql

.PHONY: postgres
postgres: build
	make -C tests/postgres/create-table run
	make -C tests/postgres/foreign-key-create run
	make -C tests/postgres/foreign-key-action run
	make -C tests/postgres/foreign-key-drop run
	make -C tests/postgres/foreign-key-alter run
	make -C tests/postgres/not-null run

.PHONY: mysql
mysql: build
	make -C tests/mysql/create-table run
	make -C tests/mysql/foreign-key-create run
	make -C tests/mysql/foreign-key-action run
	make -C tests/mysql/foreign-key-drop run
	make -C tests/mysql/foreign-key-alter run
	make -C tests/mysql/not-null run

.PHONY: build
build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	docker push $(IMAGE)
