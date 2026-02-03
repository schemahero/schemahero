#!/bin/bash

# Benchmark script for batch window feature
# This script demonstrates the performance difference when using batchWindow
#
# Prerequisites:
# - Docker installed
# - kubectl-schemahero built (make bin/kubectl-schemahero)
# - Go installed (for building plugins)
#
# Usage: ./hack/benchmark-batch-window.sh [num_tables]
# Default: 500 tables

set -e

NUM_TABLES=${1:-500}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BENCHMARK_DIR="/tmp/schemahero-benchmark-$$"
POSTGRES_CONTAINER="schemahero-benchmark-postgres"
POSTGRES_PORT=25432

echo "========================================"
echo "SchemaHero Batch Window Benchmark"
echo "========================================"
echo "Tables to create: $NUM_TABLES"
echo "Working directory: $BENCHMARK_DIR"
echo ""

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    docker rm -f $POSTGRES_CONTAINER 2>/dev/null || true
    rm -rf "$BENCHMARK_DIR" 2>/dev/null || true
}
trap cleanup EXIT

# Create benchmark directory
mkdir -p "$BENCHMARK_DIR/tables"

# Build kubectl-schemahero if needed
if [ ! -f "$ROOT_DIR/bin/kubectl-schemahero" ]; then
    echo "Building kubectl-schemahero..."
    cd "$ROOT_DIR" && make bin/kubectl-schemahero
fi

# Build postgres plugin if needed
PLUGIN_DIR="$HOME/.schemahero/plugins"
if [ ! -f "$PLUGIN_DIR/schemahero-postgres" ]; then
    echo "Building postgres plugin..."
    mkdir -p "$PLUGIN_DIR"
    cd "$ROOT_DIR/plugins/postgres" && go build -o "$PLUGIN_DIR/schemahero-postgres" .
fi

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker rm -f $POSTGRES_CONTAINER 2>/dev/null || true
docker run -d --name $POSTGRES_CONTAINER \
    -e POSTGRES_PASSWORD=benchmark \
    -e POSTGRES_USER=benchmark \
    -e POSTGRES_DB=benchmark \
    -p $POSTGRES_PORT:5432 \
    postgres:16

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker exec $POSTGRES_CONTAINER pg_isready -U benchmark -q 2>/dev/null; then
        break
    fi
    sleep 1
done
sleep 2

POSTGRES_URI="postgres://benchmark:benchmark@127.0.0.1:$POSTGRES_PORT/benchmark?sslmode=disable"

# Generate table specs
echo ""
echo "Generating $NUM_TABLES table specifications..."
for i in $(seq 1 $NUM_TABLES); do
    cat > "$BENCHMARK_DIR/tables/table_$(printf '%04d' $i).yaml" << EOF
apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: table-$(printf '%04d' $i)
spec:
  database: benchmark
  name: table_$(printf '%04d' $i)
  schema:
    postgres:
      primaryKey:
        - id
      columns:
        - name: id
          type: serial
        - name: name
          type: varchar(255)
          constraints:
            notNull: true
        - name: email
          type: varchar(255)
        - name: created_at
          type: timestamp
          default: now()
        - name: updated_at
          type: timestamp
        - name: status
          type: varchar(50)
          default: "'active'"
        - name: metadata
          type: jsonb
EOF
done

echo "Generated $NUM_TABLES table specs"
echo ""

# Benchmark 1: Individual planning (simulating no batch window)
echo "========================================"
echo "Benchmark 1: Individual Planning"
echo "(Simulates behavior WITHOUT batchWindow)"
echo "========================================"

# Reset database
docker exec $POSTGRES_CONTAINER psql -U benchmark -d benchmark -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" >/dev/null

START_TIME=$(date +%s.%N)

# Plan each table individually (this is what happens without batching)
for f in "$BENCHMARK_DIR/tables/"*.yaml; do
    "$ROOT_DIR/bin/kubectl-schemahero" plan \
        --driver postgres \
        --uri "$POSTGRES_URI" \
        --spec-file "$f" > /dev/null 2>&1
done

END_TIME=$(date +%s.%N)
INDIVIDUAL_TIME=$(echo "$END_TIME - $START_TIME" | bc)

echo "Individual planning time: ${INDIVIDUAL_TIME}s"
echo ""

# Count connections (approximate - each plan opens a connection)
echo "Estimated connections: $NUM_TABLES"
echo ""

# Benchmark 2: Batch planning (simulating batch window)
echo "========================================"
echo "Benchmark 2: Batch Planning"
echo "(Simulates behavior WITH batchWindow)"
echo "========================================"

# Reset database
docker exec $POSTGRES_CONTAINER psql -U benchmark -d benchmark -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" >/dev/null

START_TIME=$(date +%s.%N)

# Plan all tables at once (this is what batch window does)
"$ROOT_DIR/bin/kubectl-schemahero" plan \
    --driver postgres \
    --uri "$POSTGRES_URI" \
    --spec-file "$BENCHMARK_DIR/tables/" > /dev/null 2>&1

END_TIME=$(date +%s.%N)
BATCH_TIME=$(echo "$END_TIME - $START_TIME" | bc)

echo "Batch planning time: ${BATCH_TIME}s"
echo ""
echo "Estimated connections: 1"
echo ""

# Calculate improvement
SPEEDUP=$(echo "scale=2; $INDIVIDUAL_TIME / $BATCH_TIME" | bc)
SAVED_TIME=$(echo "scale=2; $INDIVIDUAL_TIME - $BATCH_TIME" | bc)
SAVED_PERCENT=$(echo "scale=1; (1 - $BATCH_TIME / $INDIVIDUAL_TIME) * 100" | bc)

echo "========================================"
echo "RESULTS"
echo "========================================"
echo ""
echo "Tables:              $NUM_TABLES"
echo ""
echo "Individual planning: ${INDIVIDUAL_TIME}s ($NUM_TABLES connections)"
echo "Batch planning:      ${BATCH_TIME}s (1 connection)"
echo ""
echo "Speedup:             ${SPEEDUP}x faster"
echo "Time saved:          ${SAVED_TIME}s (${SAVED_PERCENT}%)"
echo ""
echo "========================================"

# Output markdown for PR comment
cat > "$BENCHMARK_DIR/results.md" << EOF
## Batch Window Benchmark Results

**Configuration:**
- Tables: $NUM_TABLES
- Database: PostgreSQL 16
- Test: Planning phase only (no K8s controller overhead)

**Results:**

| Metric | Without Batch Window | With Batch Window |
|--------|---------------------|-------------------|
| Time | ${INDIVIDUAL_TIME}s | ${BATCH_TIME}s |
| DB Connections | $NUM_TABLES | 1 |

**Improvement:**
- **${SPEEDUP}x faster**
- **${SAVED_PERCENT}% time reduction**
- **$((NUM_TABLES - 1)) fewer database connections**

> Note: This benchmark measures the planning phase only. In a real K8s deployment,
> the savings would be even greater due to reduced Migration CR creation overhead
> and fewer reconciliation loops.
EOF

echo "Results saved to: $BENCHMARK_DIR/results.md"
cat "$BENCHMARK_DIR/results.md"
