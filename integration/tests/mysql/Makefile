
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-mysql-plugin 8.0.31 8.4.5 9.3.0

.PHONY: build-mysql-plugin
build-mysql-plugin:
	@echo "Building mysql plugin..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/mysql && go build -o $(HOME)/.schemahero/plugins/schemahero-mysql .



.PHONY: 8.0.31
8.0.31: export MYSQL_VERSION = 8.0.31
8.0.31: build-mysql-plugin
	make -C medium-blob run
	make -C decimal run
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
	make -C index-add run
	make -C index-drop-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run
	make -C add-collation run
	make -C remove-charset run
	make -C add-charset run
	make -C create-with-charset run
	make -C no-modify-charset run
	make -C basic-seed run
	make -C seed-with-many-rows run
	make -C multiline-seed run
	make -C multi-column-index run
	make -C primary-key-keep run
	make -C unchanged-charset run

.PHONY: 8.4.5
8.4.5: export MYSQL_VERSION = 8.4.5
8.4.5: build-mysql-plugin
	make -C medium-blob run
	make -C decimal run
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
	make -C index-add run
	make -C index-drop-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run
	make -C add-collation run
	make -C remove-charset run
	make -C add-charset run
	make -C create-with-charset run
	make -C no-modify-charset run
	make -C basic-seed run
	make -C seed-with-many-rows run
	make -C multiline-seed run
	make -C multi-column-index run
	make -C primary-key-keep run
	make -C unchanged-charset run

.PHONY: 9.3.0
9.3.0: export MYSQL_VERSION = 9.3.0
9.3.0: build-mysql-plugin
	make -C medium-blob run
	make -C decimal run
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
	make -C index-add run
	make -C index-drop-add run
	make -C primary-key-add run
	make -C primary-key-drop run
	make -C primary-key-change run
	make -C auto-increment-create run
	make -C auto-increment-add run
	make -C auto-increment-drop run
	make -C add-collation run
	make -C remove-charset run
	make -C add-charset run
	make -C create-with-charset run
	make -C no-modify-charset run
	make -C basic-seed run
	make -C seed-with-many-rows run
	make -C multiline-seed run
	make -C multi-column-index run
	make -C primary-key-keep run
	make -C unchanged-charset run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.multiarch --target --schemahero ..
	@echo "export IMAGE=$(IMAGE)"
