SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := cockroachdb
URI := postgres://schemahero@127.0.0.1:26257/schemahero?sslmode=disable

.PHONY: run
run:
	# Fixtures
	docker pull cockroachdb/cockroach:v19.2.5
    docker tag cockroachdb/cockroach:v19.2.5 $(DATABASE_IMAGE_NAME)
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run -p 26257:26257 --rm -d \
		--name $(DATABASE_CONTAINER_NAME) \
		-v `pwd`/fixtures.sql:/docker-entrypoint-initdb.d/ \
		$(DATABASE_IMAGE_NAME)
	@sleep 5
	while ! docker exec -it $(DATABASE_CONTAINER_NAME) /cockroach/cockroach sql --insecure --execute "SELECT 1;"; do sleep 1; done

	# Plan
	../../../../bin/schemahero plan --driver=$(DRIVER) --uri="$(URI)" --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	diff -B expect.sql out.sql

	# Cleanup
	@-sleep 5
	rm ./out.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
