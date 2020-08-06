SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := postgres
URI := postgres://schemahero:password@127.0.0.1:15432/schemahero?sslmode=disable

.PHONY: run
run:
	# Fixtures
	docker pull postgres:$(PG_VERSION)
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p 15432:5432 --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	while ! docker exec $(DATABASE_CONTAINER_NAME) pg_isready --quiet; do sleep 1; done
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
