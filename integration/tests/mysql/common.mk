SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := mysql
USERNAME := schemahero
PASSWORD := password
DATABASE := schemahero
URI := $(USERNAME):$(PASSWORD)@tcp(localhost:13306)/$(DATABASE)?tls=false

.PHONY: run
run:
	# Fixtures
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 || true
	docker build --build-arg MYSQL_VERSION=${MYSQL_VERSION} --no-cache -t $(DATABASE_IMAGE_NAME) .
	docker run -p 13306:3306 --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	while ! docker exec $(DATABASE_CONTAINER_NAME) mysql -u$(USERNAME) -p$(PASSWORD) $(DATABASE) -N -s -e "show tables" 2> /dev/null; do sleep 1; done
	@sleep 10

	# Plan
	../../../../bin/kubectl-schemahero plan --seed-data --driver=$(DRIVER) --uri="$(URI)" --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	if [ -f "expect-${MYSQL_VERSION}.sql" ]; then \
		if ! diff -B expect-${MYSQL_VERSION}.sql out.sql; then \
			docker logs $(DATABASE_CONTAINER_NAME); \
			exit 1; \
		fi \
	elif [ -d "expect" ] && [ -d "expect/${MYSQL_VERSION%%.*}.${MYSQL_VERSION#*.}" ] && [ -f "expect/${MYSQL_VERSION%%.*}.${MYSQL_VERSION#*.}/expect.sql" ]; then \
		if ! diff -B expect/${MYSQL_VERSION%%.*}.${MYSQL_VERSION#*.}/expect.sql out.sql; then \
			docker logs $(DATABASE_CONTAINER_NAME); \
			exit 1; \
		fi \
	elif [ -d "expect" ] && [ -d "expect/${MYSQL_VERSION%%.*}" ] && [ -f "expect/${MYSQL_VERSION%%.*}/expect.sql" ]; then \
		if ! diff -B expect/${MYSQL_VERSION%%.*}/expect.sql out.sql; then \
			docker logs $(DATABASE_CONTAINER_NAME); \
			exit 1; \
		fi \
	elif [ -f "expect.sql" ]; then \
		if ! diff -B expect.sql out.sql; then \
			docker logs $(DATABASE_CONTAINER_NAME); \
			exit 1; \
		fi \
	else \
		echo "No expect.sql file found for version ${MYSQL_VERSION}"; \
		exit 1; \
	fi

	# Apply
	if ! ../../../../bin/kubectl-schemahero apply --driver=$(DRIVER) --uri="$(URI)" --ddl out.sql; then \
		docker logs $(DATABASE_CONTAINER_NAME); \
		exit 1; \
	fi

	# Cleanup
	@-sleep 5
	# rm ./out.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
