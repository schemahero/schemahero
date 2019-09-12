SHELL := /bin/bash
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := postgres
URI := postgres://schemahero@$(DATABASE_CONTAINER_NAME):26257/schemahero?sslmode=disable

.PHONY: run
run: TEST_CONTAINER_NAME = cockroach_$(TEST_NAME)
run: NETWORK_NAME = cockroach_$(TEST_NAME)
run:
	@rm -rf ./out
	@mkdir ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	@-docker rm -f $(TEST_CONTAINER_NAME) > /dev/null 2>&1 ||:
	@-docker network rm $(NETWORK_NAME) > /dev/null 2>&1 ||:
	docker network create $(NETWORK_NAME)

	# Fixtures
	docker pull cockroachdb/cockroach:v19.1.4
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run --rm -d \
		--network $(NETWORK_NAME) \
		--name $(DATABASE_CONTAINER_NAME) \
		-v `pwd`/$(TEST_NAME)/fixtures.sql:/docker-entrypoint-initdb.d/ \
		$(DATABASE_IMAGE_NAME)
	while ! docker exec -it $(DATABASE_CONTAINER_NAME) /cockroach/cockroach sql --insecure --execute "SELECT 1;"; do sleep 1; done

	# Test
	docker tag $(IMAGE) schemahero/schemahero:test
	docker run -v `pwd`/${TEST_NAME}/spec.yaml:/specs/spec.yaml \
		--network $(NETWORK_NAME) \
		--name $(TEST_CONTAINER_NAME) \
		--rm \
		schemahero/schemahero:test \
			apply \
			--driver $(DRIVER) \
			--uri "$(URI)" \
			--spec-file /specs/spec.yaml

	# Verify
	docker run \
		--rm \
		--network $(NETWORK_NAME) \
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
	@-docker rm -f $(TEST_CONTAINER_NAME)
	@-docker network rm $(NETWORK_NAME)
