
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: v19.2.9 v20.1.4

.PHONY: v19.2.9
v19.2.9: export COCKROACHDB_VBERSION = v19.2.9
v19.2.9:
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

.PHONY: v20.1.4
v20.1.4: export COCKROACHDB_VBERSION = v20.1.4
v20.1.4:
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
