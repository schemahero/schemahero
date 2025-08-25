package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	sqlite "github.com/schemahero/schemahero/plugins/sqlite/lib"
)

// SqlitePlugin implements the DatabasePlugin interface for SQLite databases.
// It wraps the existing sqlite package functionality without duplicating code.
type SqlitePlugin struct{}

// Name returns the name of this plugin.
func (s *SqlitePlugin) Name() string {
	return "sqlite"
}

// Version returns the current version of this plugin.
func (s *SqlitePlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
func (s *SqlitePlugin) SupportedEngines() []string {
	return []string{"sqlite", "sqlite3"}
}

// Connect establishes a connection to the SQLite database using the provided URI.
// It leverages the internal sqlite.Connect function to maintain compatibility.
func (s *SqlitePlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Check if this is a fixture-only mode
	if options != nil {
		if fixtureOnly, ok := options["fixture-only"].(bool); ok && fixtureOnly {
			// Create a fixture-only connection that doesn't connect to a real database
			return sqlite.NewFixtureOnlyConnection(), nil
		}
	}

	// Use the lib sqlite package Connect function
	conn, err := sqlite.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	// Return the connection which already implements SchemaHeroDatabaseConnection
	return conn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For SQLite, we validate that required connection parameters are present.
func (s *SqlitePlugin) Validate(config map[string]interface{}) error {
	// The URI is passed separately to Connect, not in the config map
	// Config map is for additional optional parameters
	// No required parameters in config for sqlite
	return nil
}

// Initialize prepares the plugin for use.
// For SQLite, no special initialization is required as the sqlite package
// handles driver loading automatically.
func (s *SqlitePlugin) Initialize(ctx context.Context) error {
	// No initialization required for sqlite plugin
	// The sqlite package handles driver loading automatically
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For SQLite, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (s *SqlitePlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for sqlite plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}