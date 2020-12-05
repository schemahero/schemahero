SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := sqlite
URI := ./db/db.db

.PHONY: run
run:
	# Fixtures
	rm -rf ./db/db.db
	mkdir -p db
	touch ./db/db.db
	docker pull keinos/sqlite3:latest
	docker run --rm -v `pwd`/db:/db --name $(DATABASE_CONTAINER_NAME) keinos/sqlite3:latest sqlite3 /db/db.db
	docker run --rm -v `pwd`/db:/db -v `pwd`/fixtures.sql:/fixtures.sql --name $(DATABASE_CONTAINER_NAME) keinos/sqlite3:latest sqlite3 /db/db.db ".read /fixtures.sql"

	# Plan
	../../../../bin/kubectl-schemahero plan --driver=$(DRIVER) --uri="$(URI)" --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	diff -B expect.sql out.sql

	# Apply
	../../../../bin/kubectl-schemahero apply --driver=$(DRIVER) --uri="$(URI)" --ddl out.sql

	# Cleanup
	# rm ./out.sql
