
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-cassandra-plugin 3.11.10

.PHONY: build-cassandra-plugin
build-cassandra-plugin:
	@echo "Building cassandra plugin..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/cassandra && go build -o $(HOME)/.schemahero/plugins/schemahero-cassandra .

.PHONY: 3.11.10
3.11.10: export CASSANDRA_VERSION = 3.11.10
3.11.10: build-cassandra-plugin
	make -C create-table run
	make -C create-type run
	make -C add-column run
	make -C change-column-type run
	make -C create-table-with-properties run
	make -C alter-table-properties run
	make -C alter-table-default-properties-unchanged run
