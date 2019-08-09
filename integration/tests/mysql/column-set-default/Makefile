SHELL := /bin/bash
TEST_NAME := mysql-column-set-default
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := mysql
URI := schemahero:password@tcp($(DATABASE_CONTAINER_NAME):3306)/schemahero?tls=false
SPEC_FILE := /specs/users.yaml

.PHONY: run
run:
	@rm -rf ./out
	@mkdir ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	@-docker rm -f $(TEST_NAME) > /dev/null 2>&1 ||:
	@-docker network rm $(TEST_NAME) > /dev/null 2>&1 ||:
	docker network create $(TEST_NAME)

	# Fixtures
	docker pull mysql:8.0
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run --network $(TEST_NAME) --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	while ! docker exec -it $(DATABASE_CONTAINER_NAME) mysqladmin ping -h$(DATABASE_CONTAINER_NAME) --silent; do sleep 1; done

	# Test
	docker tag $(IMAGE) schemahero/schemahero:test
	docker run -v `pwd`/specs:/specs \
		--network $(TEST_NAME) \
		--name $(TEST_NAME) \
		--rm \
		schemahero/schemahero:test \
			apply \
			--driver $(DRIVER) \
			--uri "$(URI)" \
			--spec-file $(SPEC_FILE)

	# Verify
	docker run \
		--rm \
		--network $(TEST_NAME) \
		-v `pwd`/out:/out \
		-e uid=$${UID} \
		schemahero/schemahero:test \
			generate \
			--dbname schemahero \
			--namespace default \
			--driver $(DRIVER) \
			--output-dir /out \
			--uri "$(URI)"
	@echo Verifying results for $(TEST_NAME)
	diff expect out

	# Cleanup
	@-sleep 5
	rm -rf ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
	@-docker network rm $(TEST_NAME)
