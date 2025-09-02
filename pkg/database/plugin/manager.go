// Package plugin provides the interface and types for SchemaHero database plugins.
package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/schemahero/schemahero/pkg/database/interfaces"
)

// PluginManager coordinates between the plugin registry and loader, providing
// the main entry point for the plugin system. It manages plugin lifecycle
// and provides database connections via plugins.
type PluginManager struct {
	// registry manages plugin registration and metadata
	registry *PluginRegistry

	// loader handles plugin loading from various sources
	loader *PluginLoader

	// downloader handles downloading plugins from ORAS artifacts
	downloader *PluginDownloader

	// logger provides logging functionality for the manager
	logger *log.Logger
}

// NewPluginManager creates a new plugin manager with the specified registry
// and loader. The manager coordinates between these components
// to provide a unified plugin system interface.
//
// Parameters:
//
//	registry: Plugin registry for managing plugin metadata
//	loader: Plugin loader for loading plugins from sources
//
// Returns a configured PluginManager instance.
func NewPluginManager(registry *PluginRegistry, loader *PluginLoader) *PluginManager {
	if registry == nil {
		panic("registry cannot be nil")
	}
	if loader == nil {
		panic("loader cannot be nil")
	}

	return &PluginManager{
		registry:   registry,
		loader:     loader,
		downloader: NewPluginDownloader(""), // Use default cache dir
		logger:     log.New(os.Stderr, "[plugin-manager] ", log.LstdFlags),
	}
}

// GetConnection attempts to establish a database connection using a plugin for the
// specified database engine. This is the primary method for obtaining database
// connections through the plugin system.
//
// The method performs the following steps:
// 1. Checks if the plugin system is enabled
// 2. Searches for a plugin that supports the requested engine
// 3. Loads the plugin if not already loaded
// 4. Uses the plugin to establish the database connection
//
// Parameters:
//
//	ctx: Context for cancellation and timeout control
//	engine: Database engine name (e.g., "postgres", "mysql", "mongodb")
//	uri: Database connection string
//	options: Optional configuration parameters for the connection
//
// Returns a database connection interface or an error if no suitable plugin is found.
func (m *PluginManager) GetConnection(ctx context.Context, engine string, uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	if engine == "" {
		return nil, fmt.Errorf("database engine cannot be empty")
	}

	// Cassandra uses hosts/keyspace instead of URI
	if uri == "" && engine != "cassandra" {
		return nil, fmt.Errorf("database URI cannot be empty")
	}

	// Find a plugin that supports this engine
	plugin, err := m.GetDefaultPlugin(ctx, engine)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin for engine '%s': %w", engine, err)
	}

	// Validate connection options if provided
	if options != nil {
		if err := plugin.Validate(options); err != nil {
			return nil, fmt.Errorf("plugin configuration validation failed: %w", err)
		}
	}

	// Establish connection using the plugin
	conn, err := plugin.Connect(uri, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database via plugin '%s': %w", plugin.Name(), err)
	}

	return conn, nil
}

// LoadPlugin loads a plugin from the specified source and registers it in the
// registry. This method handles both local file sources and OCI registry sources
// (OCI support will be added in future PRs).
//
// Parameters:
//
//	ctx: Context for cancellation and timeout control
//	source: Plugin source configuration specifying where to load the plugin from
//
// Returns an error if the plugin cannot be loaded or registered.
func (m *PluginManager) LoadPlugin(ctx context.Context, source *PluginSource) error {
	if source == nil {
		return fmt.Errorf("plugin source cannot be nil")
	}

	var plugin DatabasePlugin
	var err error

	// Handle different source types
	switch source.Type {
	case "local":
		if source.Path == "" {
			return fmt.Errorf("local source path cannot be empty")
		}
		plugin, err = m.loader.LoadLocal(ctx, source.Path)
		if err != nil {
			return fmt.Errorf("failed to load plugin from local path '%s': %w", source.Path, err)
		}

	case "oci":
		// OCI support will be implemented in future PRs
		return fmt.Errorf("OCI source type not yet supported (coming in future PR)")

	default:
		return fmt.Errorf("unsupported plugin source type: %s", source.Type)
	}

	// Create plugin info and register it
	info := &PluginInfo{
		Name:      plugin.Name(),
		Version:   plugin.Version(),
		Engines:   plugin.SupportedEngines(),
		LocalPath: source.Path, // For local sources
	}

	if err := m.registry.Register(info); err != nil {
		// If registration fails, try to shutdown the plugin cleanly
		if shutdownErr := plugin.Shutdown(ctx); shutdownErr != nil {
			m.logger.Printf("Warning: failed to shutdown plugin after registration failure: %v", shutdownErr)
		}
		return fmt.Errorf("failed to register plugin '%s': %w", plugin.Name(), err)
	}

	return nil
}

// GetDefaultPlugin retrieves the default (first available) plugin that supports
// the specified database engine. If the plugin is not already loaded, it will
// be loaded from the registered plugin information or downloaded from ORAS artifacts.
//
// Parameters:
//
//	ctx: Context for cancellation and timeout control
//	engine: Database engine name to find a plugin for
//
// Returns the plugin that supports the engine, or an error if none is found.
func (m *PluginManager) GetDefaultPlugin(ctx context.Context, engine string) (DatabasePlugin, error) {
	if engine == "" {
		return nil, fmt.Errorf("database engine cannot be empty")
	}

	// Search for a plugin that supports this engine
	plugins := m.registry.List()
	for _, info := range plugins {
		// Check if this plugin supports the requested engine
		for _, supportedEngine := range info.Engines {
			if supportedEngine == engine {
				// Check if plugin is already loaded
				if plugin, exists := m.loader.GetPlugin(info.Name); exists {
					return plugin, nil
				}

				// Load the plugin
				plugin, err := m.loader.LoadFromInfo(ctx, info)
				if err != nil {
					// Try next plugin silently
					continue
				}

				return plugin, nil
			}
		}
	}

	// If no registered plugin found, try to download from ORAS artifacts
	return m.downloadAndLoadPlugin(ctx, engine)
}

// DiscoverPlugins searches for plugins in common locations and registers them.
// This method looks for plugin binaries in:
// - Current working directory
// - $HOME/.schemahero/plugins
// - /usr/local/lib/schemahero/plugins
// - /var/lib/schemahero/plugins
// - Directories in SCHEMAHERO_PLUGIN_PATH environment variable
func (m *PluginManager) DiscoverPlugins() {
	searchPaths := []string{
		".",             // Current directory
		"plugins/bin",   // Local development plugins directory
		"./plugins/bin", // Alternative local development plugins directory
	}

	// Add home directory plugin path FIRST (higher priority)
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, fmt.Sprintf("%s/.schemahero/plugins", homeDir))
	}

	// Add system paths LAST (lower priority)
	searchPaths = append(searchPaths,
		"/usr/local/lib/schemahero/plugins",
		"/var/lib/schemahero/plugins",
	)

	// Add paths from environment variable
	if envPaths := os.Getenv("SCHEMAHERO_PLUGIN_PATH"); envPaths != "" {
		for _, path := range filepath.SplitList(envPaths) {
			if path != "" {
				searchPaths = append(searchPaths, path)
			}
		}
	}

	// Known plugin names to search for
	knownPlugins := []string{
		"schemahero-postgres",
		"schemahero-mysql",
		"schemahero-cockroachdb",
		"schemahero-timescaledb",
		"schemahero-cassandra",
		"schemahero-sqlite",
		"schemahero-rqlite",
		"schemahero-mongodb",
	}

	for _, searchPath := range searchPaths {
		for _, pluginName := range knownPlugins {
			pluginPath := filepath.Join(searchPath, pluginName)

			// Check if file exists and is executable
			if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
				// Register the plugin without loading it yet
				// We'll extract the name from the filename (schemahero-<name>)
				baseName := filepath.Base(pluginPath)
				if len(baseName) > 11 && baseName[:11] == "schemahero-" {
					pluginName := baseName[11:] // Remove "schemahero-" prefix

					// Map plugin names to supported engines
					var engines []string
					switch pluginName {
					case "postgres":
						// Both forms supported for backward compatibility
						engines = []string{"postgres", "postgresql"}
					case "mysql":
						engines = []string{"mysql", "mariadb"}
					case "cockroachdb":
						engines = []string{"cockroachdb"}
					case "timescaledb":
						engines = []string{"timescaledb"}
					default:
						engines = []string{pluginName}
					}

					// For now, we'll register with minimal info
					// The actual plugin info will be loaded when needed
					pluginInfo := &PluginInfo{
						Name:      pluginName,
						Version:   "unknown", // Will be updated when loaded
						Engines:   engines,
						LocalPath: pluginPath,
					}

					// Check if plugin is already registered - don't overwrite
					existingPlugins := m.registry.List()
					alreadyRegistered := false
					for _, existing := range existingPlugins {
						if existing.Name == pluginName {
							alreadyRegistered = true
							break
						}
					}

					if !alreadyRegistered {
						if err := m.RegisterPlugin(pluginInfo); err != nil {
							m.logger.Printf("Failed to register plugin %s: %v", pluginPath, err)
						}
					}
				}
			}
		}
	}
}

// ListPlugins returns a list of all registered plugins. This includes both
// loaded and unloaded plugins that are registered in the plugin registry.
//
// Returns a slice of plugin information for all registered plugins.
func (m *PluginManager) ListPlugins() []*PluginInfo {
	return m.registry.List()
}

// RegisterPlugin registers a plugin in the registry using the provided plugin
// information. This does not load the plugin - it only makes it available
// for loading when needed.
//
// Parameters:
//
//	info: Plugin information to register
//
// Returns an error if the plugin information is invalid or registration fails.
func (m *PluginManager) RegisterPlugin(info *PluginInfo) error {
	if info == nil {
		return fmt.Errorf("plugin info cannot be nil")
	}

	if err := m.registry.Register(info); err != nil {
		return fmt.Errorf("failed to register plugin '%s': %w", info.Name, err)
	}

	return nil
}

// downloadAndLoadPlugin attempts to download and load a plugin for the specified engine
// from ORAS artifacts on Docker Hub using the naming convention schemahero/plugin-{engine}:{major}
func (m *PluginManager) downloadAndLoadPlugin(ctx context.Context, engine string) (DatabasePlugin, error) {
	// Get the current major version for SchemaHero
	// For now, we'll use "0" as the major version since we're in 0.x.y releases
	majorVersion := m.getCurrentMajorVersion()

	// Normalize engine name - handle aliases
	normalizedEngine := m.normalizeEngineForDownload(engine)

	// Check if plugin is already cached
	if m.downloader.IsPluginCached(normalizedEngine, majorVersion) {
		pluginPath := m.downloader.GetCachedPluginPath(normalizedEngine, majorVersion)
		return m.loadPluginFromPath(ctx, normalizedEngine, pluginPath)
	}

	// Download the plugin
	pluginPath, err := m.downloader.DownloadPlugin(ctx, normalizedEngine, majorVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to download plugin for engine '%s': %w", engine, err)
	}

	// Load the downloaded plugin
	return m.loadPluginFromPath(ctx, normalizedEngine, pluginPath)
}

// loadPluginFromPath loads a plugin from a filesystem path and registers it
func (m *PluginManager) loadPluginFromPath(ctx context.Context, engineName string, pluginPath string) (DatabasePlugin, error) {
	// Load the plugin using the loader
	plugin, err := m.loader.LoadLocal(ctx, pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin from %s: %w", pluginPath, err)
	}

	// Create plugin info and register it for future use
	info := &PluginInfo{
		Name:      plugin.Name(),
		Version:   plugin.Version(),
		Engines:   plugin.SupportedEngines(),
		LocalPath: pluginPath,
	}

	// Register the plugin (ignore errors if already registered)
	if err := m.registry.Register(info); err != nil {
		m.logger.Printf("Warning: failed to register downloaded plugin '%s': %v", plugin.Name(), err)
		// Continue anyway since the plugin is loaded
	}

	return plugin, nil
}

// getCurrentMajorVersion returns the major version for plugin compatibility
func (m *PluginManager) getCurrentMajorVersion() string {
	// Use plugin tag override if set
	if pluginTagOverride != "" {
		return pluginTagOverride
	}
	// For production builds, use major version "0" for 0.x.y releases
	return "0"
}

// normalizeEngineForDownload normalizes engine names for consistent ORAS artifact naming
func (m *PluginManager) normalizeEngineForDownload(engine string) string {
	switch strings.ToLower(engine) {
	case "postgresql":
		return "postgres"
	case "mariadb":
		return "mysql"
	default:
		return strings.ToLower(engine)
	}
}
