
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 2.3.1.0

.PHONY: 2.3.1.0
2.3.1.0: export YUGABYTEDB_VERSION = 2.3.1.0-b15
2.3.1.0:
	make -C create-table run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
