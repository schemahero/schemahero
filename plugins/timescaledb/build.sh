#!/bin/bash

set -e

echo "Building TimescaleDB plugin..."
go build -o ../bin/schemahero-timescaledb .

echo "TimescaleDB plugin built successfully: ../bin/schemahero-timescaledb"

# Optional: Check if the binary was created
if [ -f "../bin/schemahero-timescaledb" ]; then
    echo "✓ TimescaleDB plugin binary created"
    ls -lh ../bin/schemahero-timescaledb
else
    echo "✗ Failed to create TimescaleDB plugin binary"
    exit 1
fi