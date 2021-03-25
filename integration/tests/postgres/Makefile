
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 10.16 11.11 12.6 13.2

.PHONY: 10.16
10.16: export PG_VERSION = 10.16
10.16:
	make -C alter-column-timezone run
	# make -C column-set-default run
	# make -C column-unset-default run
	# make -C create-table run
	# make -C foreign-key-create run
	# make -C foreign-key-action run
	# make -C foreign-key-drop run
	# make -C foreign-key-alter run
	# make -C not-null run
	# make -C not-null-with-default run
	# make -C index-create run
	# make -C primary-key-add run
	# make -C primary-key-drop run
	# make -C unique-constraint-add run
	# make -C unique-constraint-drop run

.PHONY: 11.11
11.11: export PG_VERSION = 11.11
11.11:
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

.PHONY: 12.6
12.6: export PG_VERSION = 12.6
12.6:
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

.PHONY: 13.2
13.2: export PG_VERSION = 13.2
13.2:
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
