SHELL := /bin/bash
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := cassandra

.PHONY: run
run:
	# Fixtures
	docker pull cassandra:$(CASSANDRA_VERSION)
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p 15432:5432 --rm -d --name $(DATABASE_CONTAINER_NAME) -p 9042:9042 -v `pwd`/$(FIXTURES_FILE):/fixtures.sql cassandra:$(CASSANDRA_VERSION)
	@-docker exec $(DATABASE_CONTAINER_NAME) /bin/bash -c "while ! cqlsh -e 'describe cluster'; do sleep 1; done" > /dev/null 2>&1
	docker exec $(DATABASE_CONTAINER_NAME) cqlsh -e "CREATE KEYSPACE schemahero with replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 }"

	docker exec $(DATABASE_CONTAINER_NAME) cqlsh -k schemahero -f /fixtures.sql

	# Plan
	../../../../bin/kubectl-schemahero plan --seed-data --keyspace schemahero --host 127.0.0.1:9042 --driver=$(DRIVER) --spec-type $(SPEC_TYPE) --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	diff -B expect.sql out.sql

	# Apply
	../../../../bin/kubectl-schemahero apply --keyspace schemahero --host 127.0.0.1:9042 --driver=$(DRIVER) --ddl out.sql

	# Cleanup
	@-sleep 5
	rm ./out.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
