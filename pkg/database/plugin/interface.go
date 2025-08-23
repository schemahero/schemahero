// Package plugin provides the interface and types for SchemaHero database plugins.
// Database plugins allow extending SchemaHero to support additional database engines
// beyond the built-in ones. Plugins implement the DatabasePlugin interface and are
// responsible for managing connections to their respective database engines.
package plugin

import (
	"context"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
)

// DatabasePlugin defines the interface that all SchemaHero database plugins must implement.
// This interface provides a standardized way for SchemaHero to interact with different
// database engines through plugins, enabling extensibility and modularity.
type DatabasePlugin interface {
	// Name returns the name of the plugin. This should be a unique identifier
	// that distinguishes this plugin from others (e.g., "mongodb", "dynamodb").
	Name() string

	// Version returns the version of the plugin. This follows semantic versioning
	// conventions and is used for compatibility checking and plugin management.
	Version() string

	// SupportedEngines returns a list of database engine names that this plugin
	// supports. Multiple engines can be supported by a single plugin if they
	// share similar characteristics or protocols.
	SupportedEngines() []string

	// Connect establishes a connection to the database using the provided URI
	// and optional configuration parameters. The returned connection implements
	// the SchemaHeroDatabaseConnection interface for database operations.
	//
	// Parameters:
	//   uri: Database connection string (format depends on the database engine)
	//   options: Optional configuration parameters specific to the database engine
	//
	// Returns a connection interface or an error if the connection fails.
	Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error)

	// Validate checks the provided configuration parameters for correctness
	// and completeness. This method should verify that all required parameters
	// are present and valid before attempting to establish connections.
	//
	// Parameters:
	//   config: Configuration parameters to validate
	//
	// Returns an error if the configuration is invalid, nil otherwise.
	Validate(config map[string]interface{}) error

	// Initialize prepares the plugin for use. This method is called once during
	// plugin loading and should perform any necessary setup operations such as
	// loading drivers, initializing connection pools, or setting up logging.
	//
	// Parameters:
	//   ctx: Context for cancellation and timeout control
	//
	// Returns an error if initialization fails.
	Initialize(ctx context.Context) error

	// Shutdown cleanly terminates the plugin and releases any resources.
	// This method should close active connections, clean up temporary files,
	// and perform any other necessary cleanup operations.
	//
	// Parameters:
	//   ctx: Context for cancellation and timeout control
	//
	// Returns an error if shutdown encounters problems.
	Shutdown(ctx context.Context) error
}
