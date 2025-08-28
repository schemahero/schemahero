package main

import (
	"context"
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
	cassandra "github.com/schemahero/schemahero/plugins/cassandra/lib"
)

// CassandraPlugin implements the DatabasePlugin interface for Apache Cassandra databases.
// It wraps the existing cassandra package functionality without duplicating code.
type CassandraPlugin struct{}

// Name returns the name of this plugin.
func (p *CassandraPlugin) Name() string {
	return "cassandra"
}

// Version returns the current version of this plugin.
func (p *CassandraPlugin) Version() string {
	return "1.0.0"
}

// SupportedEngines returns the list of database engines this plugin supports.
// Only "cassandra" is supported.
func (p *CassandraPlugin) SupportedEngines() []string {
	return []string{"cassandra"}
}

// Connect establishes a connection to the Cassandra database using the provided URI.
// Note: Cassandra has a different connection model compared to SQL databases.
// The URI should contain connection information but we also need hosts, keyspace, etc.
func (p *CassandraPlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Check if this is a fixture-only mode
	if options != nil {
		if fixtureOnly, ok := options["fixture-only"].(bool); ok && fixtureOnly {
			// Create a fixture-only connection that doesn't connect to a real database
			return cassandra.NewFixtureOnlyConnection(), nil
		}
	}

	// For Cassandra, we need to parse connection parameters from options
	// Since it doesn't use a standard URI format like SQL databases
	
	// Extract hosts - required
	hostsInterface, hasHosts := options["hosts"]
	if !hasHosts {
		return nil, fmt.Errorf("hosts parameter is required for cassandra connections")
	}
	
	hosts, ok := hostsInterface.([]string)
	if !ok {
		// Try to handle single host as string
		if hostStr, ok := hostsInterface.(string); ok {
			hosts = []string{hostStr}
		} else {
			return nil, fmt.Errorf("hosts parameter must be a string array or single string")
		}
	}
	
	// Extract keyspace - required
	keyspaceInterface, hasKeyspace := options["keyspace"]
	if !hasKeyspace {
		return nil, fmt.Errorf("keyspace parameter is required for cassandra connections")
	}
	
	keyspace, ok := keyspaceInterface.(string)
	if !ok {
		return nil, fmt.Errorf("keyspace parameter must be a string")
	}
	
	// Extract optional credentials
	var username, password string
	if usernameInterface, hasUsername := options["username"]; hasUsername {
		if usernameStr, ok := usernameInterface.(string); ok {
			username = usernameStr
		}
	}
	
	if passwordInterface, hasPassword := options["password"]; hasPassword {
		if passwordStr, ok := passwordInterface.(string); ok {
			password = passwordStr
		}
	}
	
	// Use the lib cassandra package Connect function
	conn, err := cassandra.Connect(hosts, username, password, keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cassandra: %w", err)
	}

	// Return the connection which implements SchemaHeroDatabaseConnection
	return conn, nil
}

// Validate checks the provided configuration parameters for correctness.
// For Cassandra, we validate that required connection parameters are present.
func (p *CassandraPlugin) Validate(config map[string]interface{}) error {
	// Check for required hosts parameter
	if hosts, exists := config["hosts"]; !exists || hosts == nil {
		return fmt.Errorf("hosts parameter is required for cassandra connections")
	}
	
	// Check for required keyspace parameter
	if keyspace, exists := config["keyspace"]; !exists || keyspace == "" {
		return fmt.Errorf("keyspace parameter is required for cassandra connections")
	}

	return nil
}

// Initialize prepares the plugin for use.
// For Cassandra, no special initialization is required.
func (p *CassandraPlugin) Initialize(ctx context.Context) error {
	// No initialization required for cassandra plugin
	return nil
}

// Shutdown cleanly terminates the plugin and releases any resources.
// For Cassandra, no special cleanup is required at the plugin level
// as individual connections handle their own cleanup.
func (p *CassandraPlugin) Shutdown(ctx context.Context) error {
	// No shutdown required for cassandra plugin
	// Individual connections handle their own cleanup via Close()
	return nil
}