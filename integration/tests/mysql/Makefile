
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 5.6.49 5.7.31 8.0.21

.PHONY: 5.6.49
5.6.49: export MYSQL_VERSION = 5.6.49
5.6.49:
	make -C decimal run
	make -C medium-text run
	make -C medium-blob run
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C index-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run

.PHONY: 5.7.31
5.7.31: export MYSQL_VERSION = 5.7.31
5.7.31:
	make -C decimal run
	make -C medium-text run
	make -C medium-blob run
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C index-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run

.PHONY: 8.0.21
8.0.21: export MYSQL_VERSION = 8.0.21
8.0.21:
	make -C decimal run
	make -C medium-text run
	make -C medium-blob run
	make -C column-set-default run
	make -C column-unset-default run
	make -C create-table run
	make -C foreign-key-create run
	make -C foreign-key-action run
	make -C foreign-key-drop run
	make -C foreign-key-alter run
	make -C not-null run
	make -C index-create run
	make -C index-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run


.PHONY: 8.0
8.0: export MYSQL_VERSION = 8.0
8.0:
	make -C decimal run
	make -C medium-text run
	make -C medium-blob run
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
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
