#!/bin/bash

# check-plugin-leaks.sh
# This script checks for leaked SchemaHero plugin processes after test execution.
# It should be called after each integration test to ensure proper cleanup.

set -e

# Get list of all running schemahero plugin processes
# Pattern matches: schemahero-postgres, schemahero-mysql, schemahero-cockroachdb, etc.
leaked_processes=$(ps aux | grep -E "schemahero-[a-z]+" || true)

if [ -n "$leaked_processes" ]; then
    echo ""
    echo "=========================================="
    echo "ERROR: Leaked plugin processes detected!"
    echo "=========================================="
    echo ""
    echo "The following plugin processes are still running after test completion:"
    echo ""
    echo "$leaked_processes"
    echo ""
    echo "This indicates that the plugin system is not properly cleaning up"
    echo "plugin processes when the application exits."
    echo ""
    echo "Expected behavior: All plugin processes should be terminated when"
    echo "kubectl-schemahero or the manager exits."
    echo ""

    # Kill leaked processes to prevent accumulation across tests
    echo "Cleaning up leaked processes..."
    echo "$leaked_processes" | awk '{print $2}' | xargs kill -9 2>/dev/null || true

    echo ""
    echo "Test FAILED due to plugin process leak."
    exit 1
fi

echo "✓ No plugin process leaks detected"
exit 0
