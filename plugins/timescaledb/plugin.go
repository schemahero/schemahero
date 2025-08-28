package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	postgres "github.com/schemahero/schemahero/plugins/postgres/lib"
)

// TimescaleDBPlugin implements the DatabasePlugin interface for TimescaleDB databases.
// It embeds the postgres implementation and adds TimescaleDB-specific functionality.
type TimescaleDBPlugin struct{}

// Name returns the name of this plugin.
func (p *TimescaleDBPlugin) Name() string {
	return "timescaledb"
}

// Version returns the current version of this plugin.
func (p *TimescaleDBPlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
func (p *TimescaleDBPlugin) SupportedEngines() []string {
	return []string{"timescaledb"}
}

// Connect establishes a connection to the TimescaleDB database using the provided URI.
// Since TimescaleDB is PostgreSQL-based, we use the embedded postgres connection.
func (p *TimescaleDBPlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Check if this is a fixture-only mode
	if options != nil {
		if fixtureOnly, ok := options["fixture-only"].(bool); ok && fixtureOnly {
			// Create a fixture-only connection that doesn't connect to a real database
			return NewFixtureOnlyTimescaleDBConnection(), nil
		}
	}

	// Use the postgres lib package Connect function
	pgConn, err := postgres.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to timescaledb: %w", err)
	}

	// Wrap the postgres connection with TimescaleDB-specific functionality
	tsConn := NewTimescaleDBConnection(pgConn, uri)
	
	return tsConn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For TimescaleDB, we validate that required connection parameters are present.
func (p *TimescaleDBPlugin) Validate(config map[string]interface{}) error {
	// The URI is passed separately to Connect, not in the config map
	// Config map is for additional optional parameters
	// No required parameters in config for timescaledb
	return nil
}

// Initialize prepares the plugin for use.
// For TimescaleDB, no special initialization is required as the postgres package
// handles driver loading automatically.
func (p *TimescaleDBPlugin) Initialize(ctx context.Context) error {
	// No initialization required for timescaledb plugin
	// The postgres package handles driver loading automatically
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For TimescaleDB, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (p *TimescaleDBPlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for timescaledb plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}