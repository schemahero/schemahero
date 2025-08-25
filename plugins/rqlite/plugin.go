package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	rqlite "github.com/schemahero/schemahero/plugins/rqlite/lib"
)

// RqlitePlugin implements the DatabasePlugin interface for RQLite databases.
// It wraps the existing rqlite package functionality without duplicating code.
type RqlitePlugin struct{}

// Name returns the name of this plugin.
func (r *RqlitePlugin) Name() string {
	return "rqlite"
}

// Version returns the current version of this plugin.
func (r *RqlitePlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
func (r *RqlitePlugin) SupportedEngines() []string {
	return []string{"rqlite"}
}

// Connect establishes a connection to the RQLite database using the provided URI.
// It leverages the internal rqlite.Connect function to maintain compatibility.
func (r *RqlitePlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Use the lib rqlite package Connect function
	conn, err := rqlite.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rqlite: %w", err)
	}

	// Return the connection which already implements SchemaHeroDatabaseConnection
	return conn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For RQLite, we validate that required connection parameters are present.
func (r *RqlitePlugin) Validate(config map[string]interface{}) error {
	// The URI is passed separately to Connect, not in the config map
	// Config map is for additional optional parameters
	// No required parameters in config for rqlite
	return nil
}

// Initialize prepares the plugin for use.
// For RQLite, no special initialization is required as the rqlite package
// handles connection setup automatically.
func (r *RqlitePlugin) Initialize(ctx context.Context) error {
	// No initialization required for rqlite plugin
	// The rqlite package handles connection setup automatically
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For RQLite, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (r *RqlitePlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for rqlite plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}