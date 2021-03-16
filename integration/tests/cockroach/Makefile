
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: v19.2.12 v20.1.13 v20.2.6

.PHONY: v19.2.12
v19.2.2: export COCKROACHDB_VBERSION = v19.2.12
v19.2.2:
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

.PHONY: v20.1.13
v20.1.13: export COCKROACHDB_VBERSION = v20.1.13
v20.1.13:
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

.PHONY: v20.2.6
v20.2.6: export COCKROACHDB_VBERSION = v20.2.6
v20.2.6:
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
