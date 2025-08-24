#!/bin/bash

set -e

echo "Building MySQL plugin..."
go build -o ../bin/schemahero-mysql .

echo "MySQL plugin built successfully: ../bin/schemahero-mysql"

# Optional: Check if the binary was created
if [ -f "../bin/schemahero-mysql" ]; then
    echo "✓ MySQL plugin binary created"
    ls -lh ../bin/schemahero-mysql
else
    echo "✗ Failed to create MySQL plugin binary"
    exit 1
fi