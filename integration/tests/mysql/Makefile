
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 5.6 5.7 8.0

.PHONY: 5.6
5.6: export MYSQL_VERSION = 5.6
5.6:
	make -C medium-text run
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

.PHONY: 5.7
5.7: export MYSQL_VERSION = 5.7
5.7:
	make -C medium-text run
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

.PHONY: 8.0
8.0: export MYSQL_VERSION = 8.0
8.0:
	make -C medium-text run
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


.PHONY: 8.0
8.0: export MYSQL_VERSION = 8.0
8.0:
	make -C medium-text run
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
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
