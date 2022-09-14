#!/bin/bash

set -e

DATABASE_URI=$1
DATABASE_CONTAINER_NAME=$2
RQLITE_VERSION=$3

# Fixtures
docker pull rqlite/rqlite:"$RQLITE_VERSION"
docker rm -f "$DATABASE_CONTAINER_NAME" > /dev/null 2>&1 ||:
docker run --rm -v "$(pwd)"/../rqlite-auth-config.json:/rqlite/auth-config.json -p 14001:4001 -d --name "$DATABASE_CONTAINER_NAME" rqlite/rqlite:"$RQLITE_VERSION" -auth=/rqlite/auth-config.json -http-adv-addr="localhost:14001"

status_code=
while [ "$status_code" != "200" ];
do
  sleep 1;
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$DATABASE_URI"/readyz)
done

curl -s -o /dev/null "$DATABASE_URI"/db/load -H "Content-type: application/octet-stream" --data-binary @fixtures.sql
echo ""
