SHELL := /bin/bash

DATABASE_NAME ?= schemahero
DRIVER ?= postgres
INPUT_DIR ?= ./specs
OUTPUT_DIR ?= .

TEST_NAME := fixtures

.PHONY: run
run:
	@echo "Running fixtures test $(TEST_NAME) for $(DRIVER)"

	# Fixtures
	../../../../../bin/kubectl-schemahero fixtures --dbname $(DATABASE_NAME) --driver $(DRIVER) --input-dir $(INPUT_DIR) --output-dir $(OUTPUT_DIR)

	# Verify
	@echo "Verifying results for fixtures test $(TEST_NAME) for $(DRIVER)"
	diff -B $(OUTPUT_DIR)/expect.sql $(OUTPUT_DIR)/fixtures.sql

	# Cleanup
	rm $(OUTPUT_DIR)/fixtures.sql

