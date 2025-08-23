package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	postgres "github.com/schemahero/schemahero/plugins/postgres/lib"
)

// PostgresPlugin implements the DatabasePlugin interface for PostgreSQL databases.
// It wraps the existing postgres package functionality without duplicating code.
type PostgresPlugin struct{}

// Name returns the name of this plugin.
func (p *PostgresPlugin) Name() string {
	return "postgres"
}

// Version returns the current version of this plugin.
func (p *PostgresPlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
// Both "postgres" and "postgresql" are supported for backward compatibility.
// "postgres" is the preferred/modern form, "postgresql" is legacy.
func (p *PostgresPlugin) SupportedEngines() []string {
	return []string{"postgres", "postgresql"}
}

// Connect establishes a connection to the PostgreSQL database using the provided URI.
// It leverages the internal postgres.Connect function to maintain compatibility.
func (p *PostgresPlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Use the lib postgres package Connect function
	conn, err := postgres.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Return the connection which already implements SchemaHeroDatabaseConnection
	return conn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For PostgreSQL, we validate that required connection parameters are present.
func (p *PostgresPlugin) Validate(config map[string]interface{}) error {
	// Check for required URI parameter
	if uri, exists := config["uri"]; !exists || uri == "" {
		return fmt.Errorf("uri parameter is required for postgres connections")
	}

	// Additional validation could include:
	// - Checking URI format
	// - Validating optional parameters like schema, search_path, etc.
	// For now, we keep it simple and rely on the Connect method to validate the URI

	return nil
}

// Initialize prepares the plugin for use.
// For PostgreSQL, no special initialization is required as the postgres package
// handles driver loading automatically.
func (p *PostgresPlugin) Initialize(ctx context.Context) error {
	// No initialization required for postgres plugin
	// The postgres package handles driver loading automatically
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For PostgreSQL, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (p *PostgresPlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for postgres plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}