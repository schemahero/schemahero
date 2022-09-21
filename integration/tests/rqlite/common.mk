SHELL := /bin/bash
DATABASE_CONTAINER_NAME := schemahero-database
DRIVER := rqlite
URI := http://schemahero:notasecret@localhost:14001

.PHONY: run
run:
	../bootstrap.sh $(URI) $(DATABASE_CONTAINER_NAME) $(RQLITE_VERSION)

	# Plan
	../../../../bin/kubectl-schemahero plan --seed-data --driver=$(DRIVER) --uri="$(URI)" --spec-file $(SPEC_FILE) > out.sql

	# Verify
	@echo Verifying results for $(TEST_NAME)
	diff -B expect.sql out.sql

	# Apply
	../../../../bin/kubectl-schemahero apply --driver=$(DRIVER) --uri="$(URI)" --ddl out.sql

	# Cleanup
	@-sleep 5
	rm ./out.sql
	@-docker rm -f $(DATABASE_CONTAINER_NAME)
