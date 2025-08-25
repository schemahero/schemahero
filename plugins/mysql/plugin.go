package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	mysql "github.com/schemahero/schemahero/plugins/mysql/lib"
)

// MySQLPlugin implements the DatabasePlugin interface for MySQL databases.
// It wraps the existing mysql package functionality without duplicating code.
type MySQLPlugin struct{}

// Name returns the name of this plugin.
func (p *MySQLPlugin) Name() string {
	return "mysql"
}

// Version returns the current version of this plugin.
func (p *MySQLPlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
// This includes MySQL and MariaDB which share the same protocol.
func (p *MySQLPlugin) SupportedEngines() []string {
	return []string{"mysql", "mariadb"}
}

// Connect establishes a connection to the MySQL database using the provided URI.
// It leverages the internal mysql.Connect function to maintain compatibility.
func (p *MySQLPlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Use the internal mysql package Connect function
	conn, err := mysql.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	// Return the connection which already implements SchemaHeroDatabaseConnection
	return conn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For MySQL, we validate that required connection parameters are present.
func (p *MySQLPlugin) Validate(config map[string]interface{}) error {
	// The URI is passed separately to Connect, not in the config map
	// Config map is for additional optional parameters
	// No required parameters in config for mysql
	return nil
}

// Initialize prepares the plugin for use.
// For MySQL, no special initialization is required as the mysql package
// handles driver loading automatically.
func (p *MySQLPlugin) Initialize(ctx context.Context) error {
	// No initialization required for mysql plugin
	// The mysql package handles driver loading automatically
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For MySQL, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (p *MySQLPlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for mysql plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}