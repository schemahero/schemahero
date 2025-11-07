
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-rqlite-plugin 7.6.1 6.10.2

.PHONY: build-rqlite-plugin
build-rqlite-plugin:
	@echo "Building rqlite plugin..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/rqlite && go build -o $(HOME)/.schemahero/plugins/schemahero-rqlite .

.PHONY: 7.6.1
7.6.1: export RQLITE_VERSION = 7.6.1
7.6.1: build-rqlite-plugin
	make -C basic-seed run
	make -C column-add run
	make -C column-alter run
	make -C column-drop run
	make -C column-set-default run
	make -C column-unset-default run
	make -C drop-indexed-column run
	make -C drop-unique-constraint-column run
	make -C drop-unique-indexed-column run
	make -C foreign-key-action run
	make -C foreign-key-alter run
	make -C foreign-key-create run
	make -C foreign-key-drop run
	make -C index-alter run
	make -C index-create run
	make -C index-drop run
	make -C index-named-no-change run
	make -C index-no-change run
	make -C not-null run
	make -C not-null-with-default run
	make -C primary-key-add run
	make -C primary-key-alter run
	make -C primary-key-drop run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C table-create run
	make -C table-create-strict run
	make -C table-create-with-index run
	make -C table-create-with-unique-index run
	make -C unique-constraint-add run
	make -C unique-constraint-alter run
	make -C unique-constraint-drop run
	make -C unique-index-alter run
	make -C unique-index-create run
	make -C unique-index-drop run
	make -C unique-index-named-no-change run
	make -C unique-index-no-change run

.PHONY: 6.10.2
6.10.2: export RQLITE_VERSION = 6.10.2
6.10.2: build-rqlite-plugin
	make -C basic-seed run
	make -C column-add run
	make -C column-alter run
	make -C column-drop run
	make -C column-set-default run
	make -C column-unset-default run
	make -C drop-indexed-column run
	make -C drop-unique-constraint-column run
	make -C drop-unique-indexed-column run
	make -C foreign-key-action run
	make -C foreign-key-alter run
	make -C foreign-key-create run
	make -C foreign-key-drop run
	make -C index-alter run
	make -C index-create run
	make -C index-drop run
	make -C index-named-no-change run
	make -C index-no-change run
	make -C not-null run
	make -C not-null-with-default run
	make -C primary-key-add run
	make -C primary-key-alter run
	make -C primary-key-drop run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
	make -C table-create run
	make -C table-create-with-index run
	make -C table-create-with-unique-index run
	make -C unique-constraint-add run
	make -C unique-constraint-alter run
	make -C unique-constraint-drop run
	make -C unique-index-alter run
	make -C unique-index-create run
	make -C unique-index-drop run
	make -C unique-index-named-no-change run
	make -C unique-index-no-change run

.PHONY: seed
seed:
	make -C basic-seed run
	make -C seed-data-without-schema run
	make -C seed-with-many-rows run
