SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
URI := postgres://schemahero:password@127.0.0.1:5433/schemahero?sslmode=disable

.PHONY: run
run:
	# Fixtures https://github.com/yugabyte/yugabyte-db/issues/4880
	docker pull yugabytedb/yugabyte:$(YUGABYTEDB_VERSION)
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p7000:7000 -p9000:9000 -p5433:5433 -p9042:9042 --rm -d -v `pwd`/$(FIXTURES_FILE):/fixtures.sql --name $(DATABASE_CONTAINER_NAME) yugabytedb/yugabyte:$(YUGABYTEDB_VERSION) bin/yugabyted start --daemon=false
	@sleep 5
	docker exec $(DATABASE_CONTAINER_NAME) ./bin/ysqlsh -c "CREATE DATABASE schemahero"
	docker exec $(DATABASE_CONTAINER_NAME) ./bin/ysqlsh -c "CREATE USER schemahero WITH PASSWORD 'password'"
	docker exec $(DATABASE_CONTAINER_NAME) ./bin/ysqlsh -c "GRANT ALL PRIVILEGES ON DATABASE schemahero TO schemahero"
	
	docker exec $(DATABASE_CONTAINER_NAME) ./bin/ysqlsh -d schemahero -f /fixtures.sql

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
