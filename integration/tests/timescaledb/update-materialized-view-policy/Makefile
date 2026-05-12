SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := timescaledb
POSTGRES_PASSWORD_URI := %21%40%23%24%25%5E%26%2A%28%29%7B%7D%27%22%3B
URI := postgres://schemahero:$(POSTGRES_PASSWORD_URI)@127.0.0.1:15432/schemahero?sslmode=disable
TEST_NAME := timescaledb-update-materialized-view-policy

.PHONY: run
run:
	# Fixtures
	docker pull timescale/timescaledb:$(TIMESCALEDB_VERSION)
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p 15432:5432 --rm -d -e POSTGRES_PASSWORD='!@#$$%^&*(){}'\''";' -e POSTGRES_HOST_AUTH_METHOD=md5 --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	while ! docker exec $(DATABASE_CONTAINER_NAME) pg_isready --quiet; do sleep 1; done
	@sleep 10

	# Phase 1: Plan initial view with refresh policy
	../../../../bin/kubectl-schemahero plan --seed-data --driver=$(DRIVER) --uri="$(URI)" --spec-file ./specs-initial > out-initial.sql

	# Verify Phase 1
	@echo Verifying initial plan for $(TEST_NAME)
	diff -B expect-initial.sql out-initial.sql

	# Apply Phase 1
	../../../../bin/kubectl-schemahero apply --driver=$(DRIVER) --uri="$(URI)" --ddl out-initial.sql

	# Phase 2: Plan updated policy
	../../../../bin/kubectl-schemahero plan --seed-data --driver=$(DRIVER) --uri="$(URI)" --spec-file ./specs-updated > out-update.sql

	# Verify Phase 2
	@echo Verifying update plan for $(TEST_NAME)
	diff -B expect-update.sql out-update.sql

	# Check for leaked plugin processes
	@bash ../../../check-plugin-leaks.sh

	# Cleanup
	@-sleep 5
	rm ./out-initial.sql ./out-update.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
