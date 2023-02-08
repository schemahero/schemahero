
SHELL := /bin/bash

export IMAGE
export GO111MODULE=on

.PHONY: run
run: postgres mysql cockroach

.PHONY: postgres
postgres:
	make -C tests/postgres run

.PHONY: mysql
mysql:
	make -C tests/mysql run

.PHONY: cockroach
cockroach:
	make -C tests/cockroach run

.PHONY: sqlite
sqlite:
	make -C tests/sqlite run

.PHONY: rqlite
rqlite:
	make -C tests/rqlite run

.PHONY: build
build: docker-build
	docker push $(IMAGE)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) -f ../Dockerfile.multiarch --target --schemahero ..
	@echo "export IMAGE=$(IMAGE)"
