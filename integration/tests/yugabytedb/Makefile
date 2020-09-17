
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 2.1.4.0

.PHONY: 2.1.4.0
2.1.4.0: export YUGABYTEDB_VERSION = 2.1.4.0-b5
2.1.4.0:
	make -C create-table-ysql run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
