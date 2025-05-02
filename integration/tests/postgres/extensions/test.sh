set -e

export POSTGRES_USER=${POSTGRES_USER:-schemahero}
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
export POSTGRES_DB=${POSTGRES_DB:-schemahero}

docker-compose up -d

echo "Waiting for PostgreSQL to be ready..."
sleep 10

cd ../../../../
make bin/kubectl-schemahero

echo "Testing with pgvector image (extension available)..."
POSTGRES_URI="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable"

cat <<EOF > /tmp/postgres-db.yaml
apiVersion: databases.schemahero.io/v1alpha4
kind: Database
metadata:
  name: postgres
spec:
  connection:
    postgres:
      uri:
        value: ${POSTGRES_URI}
EOF

./bin/kubectl-schemahero apply -f /tmp/postgres-db.yaml

./bin/kubectl-schemahero apply -f integration/tests/postgres/extensions/postgres-extension.yaml

PGPASSWORD=${POSTGRES_PASSWORD} psql -h localhost -p 5432 -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "SELECT * FROM pg_extension WHERE extname = 'vector';" | grep -q "vector" || (echo "Extension 'vector' not found in pgvector image" && exit 1)
echo "Extension 'vector' successfully created in pgvector image"

echo "Testing with regular PostgreSQL image (extension not available)..."
POSTGRES_URI="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5433/${POSTGRES_DB}?sslmode=disable"

cat <<EOF > /tmp/postgres-no-vector-db.yaml
apiVersion: databases.schemahero.io/v1alpha4
kind: Database
metadata:
  name: postgres-no-vector
spec:
  connection:
    postgres:
      uri:
        value: ${POSTGRES_URI}
EOF

./bin/kubectl-schemahero apply -f /tmp/postgres-no-vector-db.yaml

cat <<EOF > /tmp/postgres-no-vector-extension.yaml
apiVersion: schemas.schemahero.io/v1alpha4
kind: DatabaseExtension
metadata:
  name: vector-extension-no-vector
spec:
  database: postgres-no-vector
  postgres:
    name: vector
EOF

if ./bin/kubectl-schemahero apply -f /tmp/postgres-no-vector-extension.yaml; then
  if PGPASSWORD=${POSTGRES_PASSWORD} psql -h localhost -p 5433 -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "SELECT * FROM pg_extension WHERE extname = 'vector';" | grep -q "vector"; then
    echo "Error: Extension 'vector' was created in regular PostgreSQL image, but it shouldn't be possible"
    exit 1
  else
    echo "Extension 'vector' not created in regular PostgreSQL image as expected"
  fi
else
  echo "Failed to create extension in regular PostgreSQL image as expected"
fi

docker-compose down

echo "All tests passed!"
exit 0
