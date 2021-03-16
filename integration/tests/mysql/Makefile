
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 5.6.51 5.7.33 8.0.23

.PHONY: 5.6.51
5.6.51: export MYSQL_VERSION = 5.6.51
5.6.51:
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
	make -C add-charset run
	make -C add-collation run
	make -C remove-charset run

.PHONY: 5.7.33
5.7.33: export MYSQL_VERSION = 5.7.33
5.7.33:
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
	make -C add-charset run
	make -C add-collation run
	make -C remove-charset run

.PHONY: 8.0.23
8.0.23: export MYSQL_VERSION = 8.0.23
8.0.23:
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
	make -C add-charset run
	make -C add-collation run
	make -C remove-charset run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
