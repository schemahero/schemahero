SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: build-cassandra-plugin 4.0.19 4.1.10 5.0.6

.PHONY: build-cassandra-plugin
build-cassandra-plugin:
	@echo "Building cassandra plugin..."
	@mkdir -p $(HOME)/.schemahero/plugins
	@cd ../../../plugins/cassandra && go build -o $(HOME)/.schemahero/plugins/schemahero-cassandra .

.PHONY: 4.0.19
4.0.19: export CASSANDRA_VERSION = 4.0.19
4.0.19: build-cassandra-plugin
	make -C create-table run
	make -C create-type run
	make -C add-column run
	make -C change-column-type run
	make -C create-table-with-properties run
	make -C alter-table-properties run
	make -C alter-table-default-properties-unchanged run
	make -C seed-data-without-schema run

.PHONY: 4.1.10
4.1.10: export CASSANDRA_VERSION = 4.1.10
4.1.10: build-cassandra-plugin
	make -C create-table run
	make -C create-type run
	make -C add-column run
	make -C change-column-type run
	make -C create-table-with-properties run
	make -C alter-table-properties run
	make -C alter-table-default-properties-unchanged run
	make -C seed-data-without-schema run

.PHONY: 5.0.6
5.0.6: export CASSANDRA_VERSION = 5.0.6
5.0.6: build-cassandra-plugin
	make -C create-table run
	make -C create-type run
	make -C add-column run
	make -C change-column-type run
	make -C create-table-with-properties run
	make -C alter-table-properties run
	make -C alter-table-default-properties-unchanged run
	make -C seed-data-without-schema run
