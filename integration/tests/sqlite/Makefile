
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: 3.33.0

.PHONY: 3.33.0
3.33.0: export SQLITE_VERSION = 3.33.0
3.33.0:
	make -C create-table run
	make -C not-null run
	make -C not-null-with-default run
	# make -C primary-key-add run
	# make -C primary-key-drop run
		
.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.schemahero ..
	@echo "export IMAGE=$(IMAGE)"
