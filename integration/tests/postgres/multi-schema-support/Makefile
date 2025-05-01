include ../common.mk

TEST_NAME ?= multi-schema-support
POSTGRES_TAG ?= postgres:16

.PHONY: run
run:
	@docker build -t $(TEST_NAME) -f ./Dockerfile .
	@docker run --rm -p 15432:5432 -d --name $(TEST_NAME) $(TEST_NAME)
	@sleep 5
	@../../bin/kubectl-schemahero generate --driver postgres --uri postgres://schemahero:password@localhost:15432/schemahero?sslmode=disable --dbname schemahero --schema test --schema public --output-dir ./out
	@docker stop $(TEST_NAME)
	@echo "Validating results..."
	@cat ./out/users.yaml | grep "schema: public" || (echo "Failed to find public schema in users.yaml" && exit 1)
	@cat ./out/test-users.yaml | grep "schema: test" || (echo "Failed to find test schema in test-users.yaml" && exit 1)
	@echo "Test passed!"
