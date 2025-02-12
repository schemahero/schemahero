#!/bin/bash

# build the schemahero image and push to ttl.sh
# suffix=$(date +%s)
# docker build -f ./deploy/Dockerfile.multiarch --target manager -t ttl.sh/schemahero/schemahero-e2e-${suffix}:1h .
# docker push ttl.sh/schemahero/schemahero-e2e-${suffix}:1h

docker build -f ./deploy/Dockerfile.multiarch --target manager -t ttl.sh/schemahero/schemahero-e2e-manager:720h .
docker build -f ./deploy/Dockerfile.multiarch --target schemahero -t ttl.sh/schemahero/schemahero-e2e-schemahero:720h .
docker push ttl.sh/schemahero/schemahero-e2e-manager:720h
docker push ttl.sh/schemahero/schemahero-e2e-schemahero:720h

# generate an operator yaml with that uses the locally pushed tag
rm -rf ./e2e-install
mkdir -p ./e2e-install
docker run --rm  \
    -v $(pwd)/e2e-install:/e2e-install \
    -e SCHEMAHERO_IMAGE=ttl.sh/schemahero/schemahero-e2e-manager:720h \
    -u $(id -u):$(id -g) \
    ttl.sh/schemahero/schemahero-e2e-schemahero:720h \
    install --yaml --manager-image ttl.sh/schemahero/schemahero-e2e-manager:720h --out-dir /e2e-install

# create the cluster
# /Users/marc/go/src/github.com/replicatedhq/replicated/bin/replicated cluster create --distribution kind --version v1.25.0 --name schemahero-postgres --wait 5m

# get the kubeconfig
# /Users/marc/go/src/github.com/replicatedhq/replicated/bin/replicated cluster kubeconfig --name schemahero-postgres --output kubeconfig.yaml

# install schemahero
kubectl create ns --kubeconfig ./kubeconfig.yaml schemahero-system
kubectl apply --kubeconfig ./kubeconfig.yaml -f ./e2e-install

# install postgres
# kubectl create ns --kubeconfig ./kubeconfig.yaml schemahero-e2e
# kubectl apply --kubeconfig ./kubeconfig.yaml -f ./e2e/postgres/postgres.yaml
# wait for postgres
# sleep 30

# manually create a few tables that we will alter
# kubectl --kubeconfig ./kubeconfig.yaml --namespace schemahero-e2e exec -it postgresql-0 -- psql -U postgres -c "CREATE TABLE schemahero_test (id int, name text);"
# kubectl --kubeconfig ./kubeconfig.yaml --namespace schemahero-e2e exec -it postgresql-0 -- psql -U postgres -c "CREATE TABLE schemahero_test2 (id int, name text);"

# # create a database object
kubectl --kubeconfig ./kubeconfig.yaml apply -f ./e2e/postgres/database.yaml

# # install all of the tables
kubectl --kubeconfig ./kubeconfig.yaml apply -f ./e2e/postgres/tables.yaml

# # wait for reconcile
# sleep 30

# # check that all migration objects have been created and applied
kubectl --kubeconfig ./kubeconfig.yaml get tables -n schemahero-e2e

# # verify the tables have the desired schema
# kubectl exec -it postgres-0 -- psql -U postgres -c "\d schemahero_test"
# kubectl exec -it postgres-0 -- psql -U postgres -c "\d schemahero_test2"

# delete the cluster now instead of waiting for ttl to expire
# if the script gets here, that means all tests succeeded and we don't want this cluster
# /Users/marc/go/src/github.com/replicatedhq/replicated/bin/replicated cluster delete --name schemahero-postgres