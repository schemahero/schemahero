SHELL := /bin/bash
TEST_NAME := postgres-foreign-key-drop
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := schemahero-database

.PHONY: run
run:
	@rm -rf ./out
	@mkdir ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	@-docker rm -f $(TEST_NAME) > /dev/null 2>&1 ||:
	@-docker network rm $(TEST_NAME) > /dev/null 2>&1 ||:
	docker network create $(TEST_NAME)

	# Fixtures
	docker pull postgres:10.7
	docker build -t $(DATABASE_IMAGE_NAME) .
	@-docker rm -f $(DATABASE_CONTAINER_NAME) > /dev/null 2>&1 ||:
	docker run --network $(TEST_NAME) --rm -d --name $(DATABASE_CONTAINER_NAME) $(DATABASE_IMAGE_NAME)
	@-sleep 5

	# Test
	docker tag $(IMAGE) schemahero/schemahero:test
	docker run -v `pwd`/specs:/specs \
		--network $(TEST_NAME) \
		--name $(TEST_NAME) \
		--rm \
		schemahero/schemahero:test \
			apply \
			--driver postgres \
			--uri postgres://schemahero:password@$(DATABASE_CONTAINER_NAME):5432/schemahero?sslmode=disable \
			--spec-file /specs/org.yaml

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
			--driver postgres \
			--output-dir /out \
			--uri postgres://schemahero:password@$(DATABASE_CONTAINER_NAME):5432/schemahero?sslmode=disable
	@echo Verifying results for $(TEST_NAME)
	diff expect out

	# Cleanup
	@-sleep 5
	rm -rf ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
	@-docker network rm $(TEST_NAME)

