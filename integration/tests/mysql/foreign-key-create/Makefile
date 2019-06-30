SHELL := /bin/bash
TEST_NAME := mysql-foreign-key-create
DATABASE_IMAGE_NAME := schemahero/database
DATABASE_CONTAINER_NAME := database

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
	while ! docker exec -it $(DATABASE_CONTAINER_NAME) mysqladmin ping -hlocalhost --silent; do sleep 30; done

	# Test
	docker tag $(IMAGE) schemahero/schemahero:test
	docker run -v `pwd`/specs:/specs \
		--network $(TEST_NAME) \
		--name $(TEST_NAME) \
		--rm \
		schemahero/schemahero:test \
			apply \
			--driver mysql \
			--uri "schemahero:password@tcp($(DATABASE_CONTAINER_NAME):3306)/schemahero?tls=false" \
			--spec-file /specs/user-project.yaml

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
			--driver mysql \
			--output-dir /out \
			--uri "schemahero:password@tcp($(DATABASE_CONTAINER_NAME):3306)/schemahero?tls=false"
	@echo Verifying results for $(TEST_NAME)
	diff --color expect out

	# Cleanup
	@-sleep 5
	rm -rf ./out
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
	@-docker network rm $(TEST_NAME)

