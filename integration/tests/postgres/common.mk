SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := postgres
URI := postgres://schemahero:password@$(DATABASE_CONTAINER_NAME):5432/schemahero?sslmode=disable

.PHONY: run
run:
	@rm -rf ./out
	@mkdir ./out
	@chmod 777 ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	@-docker rm -f $(TEST_NAME) > /dev/null 2>&1 ||:
	@-docker network rm $(TEST_NAME) > /dev/null 2>&1 ||:
	docker network create $(TEST_NAME)

	# Fixtures
	docker pull postgres:10.7
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run --network $(TEST_NAME) --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	while ! docker exec -it $(DATABASE_CONTAINER_NAME) pg_isready -h$(DATABASE_CONTAINER_NAME) --quiet; do sleep 1; done

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
