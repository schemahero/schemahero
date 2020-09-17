SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := yugabytedb
URI := postgres://schemahero:password@127.0.0.1:15432/schemahero?sslmode=disable

.PHONY: run
run:
	# Fixtures
	docker pull yugabytedb/yugabyte:$(YUGABYTEDB_VERSION)
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p7000:7000 -p9000:9000 -p5433:5433 -p9042:9042 --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME) start --daemon=false
	@sleep 1

	# Plan
	../../../../bin/kubectl-schemahero plan --driver=$(DRIVER) --uri="$(URI)" --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	diff -B expect.sql out.sql

	# Apply
	../../../../bin/kubectl-schemahero apply --driver=$(DRIVER) --uri="$(URI)" --ddl out.sql

	# Cleanup
	@-sleep 5
	rm ./out.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
