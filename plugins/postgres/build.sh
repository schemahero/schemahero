#!/bin/bash

# Build script for SchemaHero Postgres Plugin
set -e

PLUGIN_NAME="schemahero-postgres"
OUTPUT_DIR="${OUTPUT_DIR:-./bin}"

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"

echo "Building Postgres plugin..."

# Build the plugin binary
go build -o "$OUTPUT_DIR/$PLUGIN_NAME" .

echo "Postgres plugin built successfully: $OUTPUT_DIR/$PLUGIN_NAME"

# Make it executable
chmod +x "$OUTPUT_DIR/$PLUGIN_NAME"

echo "Plugin is ready at: $OUTPUT_DIR/$PLUGIN_NAME"