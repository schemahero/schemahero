
SHELL := /bin/bash

.PHONY: run
run: kind

.PHONY: kind
kind:
	# create the cluster
	/Users/marc/go/src/github.com/replicatedhq/replicated/bin/replicated cluster create --distribution kind --version v1.25.0 --name schemahero-postgres --wait 5m

	# get the kubeconfig
	/Users/marc/go/src/github.com/replicatedhq/replicated/bin/replicated cluster kubeconfig --name schemahero-postgres --output kubeconfig.yaml

	# build the schemahero image and push to ttl.sh
	docker build -t ttl.sh/marc/schemahero:latest .
	docker push ttl.sh/marc/schemahero:latest
	
	# install schemahero
	kubectl apply -f https://raw.githubusercontent.com/schemahero/schemahero/master/deploy/operator.yaml

	# install postgres
	kubectl apply -f https://raw.githubusercontent.com/schemahero/schemahero/master/examples/postgres/postgres.yaml

	# manually create a few tables that we will alter
	kubectl exec -it postgres-0 -- psql -U postgres -c "CREATE TABLE schemahero_test (id int, name text);"
	kubectl exec -it postgres-0 -- psql -U postgres -c "CREATE TABLE schemahero_test2 (id int, name text);"

	# create a database object
	kubectl apply -f ./database.yaml

	# install all of the tables
	kubectl apply -f ./tables.yaml

	# wait for reconcile
	kubectl wait --for=condition=ready --timeout=5m database schemahero-test

	# check that all migration objects have been created and applied
	kubectl get migrations

	# verify the tables have the desired schema
	kubectl exec -it postgres-0 -- psql -U postgres -c "\d schemahero_test"
	kubectl exec -it postgres-0 -- psql -U postgres -c "\d schemahero_test2"
