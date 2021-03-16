
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 3.11.10

.PHONY: 3.11.10
3.11.10: export CASSANDRA_VERSION = 3.11.10
3.11.10:
	make -C create-table run
	make -C create-type run
	make -C add-column run
	make -C change-column-type run
	make -C create-table-with-properties run
	make -C alter-table-properties run
	make -C alter-table-default-properties-unchanged run
