// Package plugin provides the interface and types for SchemaHero database plugins.
package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/schemahero/schemahero/pkg/database/plugin/shared"
)

// PluginLoader manages the loading and caching of database plugins.
// It provides thread-safe operations for loading plugins from various sources
// and maintains a cache of loaded plugins for efficient reuse.
type PluginLoader struct {
	// pluginDir is the directory where plugins are stored locally
	pluginDir string

	// cache stores loaded plugins by name for efficient reuse
	cache map[string]DatabasePlugin

	// clients stores plugin client processes by plugin name for cleanup
	clients map[string]*plugin.Client

	// cacheMutex protects concurrent access to the plugin cache and clients map
	cacheMutex sync.RWMutex

	// logger provides logging functionality for the loader
	logger *log.Logger
}

// NewPluginLoader creates a new plugin loader with the specified plugin directory.
// If pluginDir is empty, it defaults to "/var/lib/schemahero/plugins".
//
// Parameters:
//
//	pluginDir: Directory path where plugins are stored locally
//
// Returns a configured PluginLoader instance.
func NewPluginLoader(pluginDir string) *PluginLoader {
	if pluginDir == "" {
		pluginDir = "/var/lib/schemahero/plugins"
	}

	// Initialize the plugin map to avoid circular dependencies
	shared.PluginMap[shared.DatabaseInterfaceName] = &DatabaseRPCPlugin{}

	return &PluginLoader{
		pluginDir:  pluginDir,
		cache:      make(map[string]DatabasePlugin),
		clients:    make(map[string]*plugin.Client),
		cacheMutex: sync.RWMutex{},
		logger:     log.New(os.Stderr, "[plugin-loader] ", log.LstdFlags),
	}
}

// LoadLocal loads a plugin from a local filesystem path.
// It validates that the file exists and is executable before attempting to load it.
// The loaded plugin is cached for future use.
//
// Parameters:
//
//	ctx: Context for cancellation and timeout control
//	path: Filesystem path to the plugin binary
//
// Returns the loaded plugin or an error if loading fails.
func (l *PluginLoader) LoadLocal(ctx context.Context, path string) (DatabasePlugin, error) {
	if path == "" {
		return nil, fmt.Errorf("plugin path cannot be empty")
	}

	// Check if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin file does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to stat plugin file %s: %w", path, err)
	}

	// Check if file is a regular file (not a directory or symlink)
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("plugin path is not a regular file: %s", path)
	}

	// Check if file is executable
	if fileInfo.Mode().Perm()&0111 == 0 {
		return nil, fmt.Errorf("plugin file is not executable: %s", path)
	}

	// Loading plugin from local path
	l.logger.Printf("Loading plugin from local path: %s", path)

	// Create the plugin command
	cmd := exec.Command(path)

	// Create a logger that will capture plugin stderr output
	pluginLogger := hclog.New(&hclog.LoggerOptions{
		Name:   fmt.Sprintf("plugin-%s", fileInfo.Name()),
		Output: os.Stderr,
		Level:  hclog.Debug,
	})

	// Create the plugin client with logging enabled to capture errors
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             cmd,
		Logger:          pluginLogger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
		SyncStdout: os.Stdout,
		SyncStderr: os.Stderr,
	})

	l.logger.Printf("Starting plugin client for: %s", path)
	// Start the client
	rpcClient, err := client.Client()
	if err != nil {
		l.logger.Printf("ERROR: Failed to start plugin client for %s: %v", path, err)
		client.Kill()
		return nil, fmt.Errorf("failed to start plugin client for %s: %w", path, err)
	}
	l.logger.Printf("Plugin client started successfully for: %s", path)

	// Get the plugin interface
	raw, err := rpcClient.Dispense(shared.DatabaseInterfaceName)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin from %s: %w", path, err)
	}

	// Cast to our plugin interface
	plugin, ok := raw.(DatabasePlugin)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin from %s does not implement DatabasePlugin interface", path)
	}

	// Initialize the plugin
	if err := plugin.Initialize(ctx); err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to initialize plugin from %s: %w", path, err)
	}

	// Validate the plugin
	if err := l.ValidatePlugin(plugin); err != nil {
		return nil, fmt.Errorf("plugin validation failed for %s: %w", path, err)
	}

	// Cache the plugin and store the client for cleanup
	l.cacheMutex.Lock()
	l.cache[plugin.Name()] = plugin
	l.clients[plugin.Name()] = client
	l.cacheMutex.Unlock()

	// Successfully loaded plugin

	return plugin, nil
}

// LoadFromInfo loads a plugin based on the provided PluginInfo.
// For PR 1, only LocalPath is supported. Future PRs will add OCI registry support.
//
// Parameters:
//
//	ctx: Context for cancellation and timeout control
//	info: Plugin information containing loading details
//
// Returns the loaded plugin or an error if loading fails.
func (l *PluginLoader) LoadFromInfo(ctx context.Context, info *PluginInfo) (DatabasePlugin, error) {
	if info == nil {
		return nil, fmt.Errorf("plugin info cannot be nil")
	}

	if info.Name == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	// Check if plugin is already cached
	if plugin, exists := l.GetPlugin(info.Name); exists {
		// Using cached plugin
		return plugin, nil
	}

	// For PR 1, only support LocalPath loading
	if info.LocalPath == "" {
		return nil, fmt.Errorf("local path is required for plugin loading (OCI support coming in future PR)")
	}

	// Loading plugin from info

	return l.LoadLocal(ctx, info.LocalPath)
}

// Unload removes a plugin from the cache and shuts it down cleanly.
// This allows for plugin updates and memory cleanup.
//
// Parameters:
//
//	name: Name of the plugin to unload
//
// Returns an error if the plugin is not found or if shutdown fails.
func (l *PluginLoader) Unload(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	plugin, exists := l.cache[name]
	if !exists {
		return fmt.Errorf("plugin '%s' is not loaded", name)
	}

	// Shutdown the plugin cleanly
	ctx := context.Background()
	if err := plugin.Shutdown(ctx); err != nil {
		l.logger.Printf("Warning: failed to shutdown plugin %s: %v", name, err)
		// Continue with unloading even if shutdown fails
	}

	// Kill the plugin client process if it exists
	if client, exists := l.clients[name]; exists {
		client.Kill()
		delete(l.clients, name)
	}

	// Remove from cache
	delete(l.cache, name)

	// Successfully unloaded plugin

	return nil
}

// Cleanup kills all plugin client processes and clears the cache.
// This method should be called when shutting down the plugin loader
// to ensure all plugin processes are properly terminated.
func (l *PluginLoader) Cleanup() {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	// Kill all plugin clients
	for _, client := range l.clients {
		// Killing plugin client
		client.Kill()
	}

	// Clear all caches
	l.cache = make(map[string]DatabasePlugin)
	l.clients = make(map[string]*plugin.Client)

	// Plugin loader cleanup completed
}

// GetPlugin retrieves a cached plugin by name.
// This method is thread-safe and returns both the plugin and a boolean
// indicating whether the plugin was found in the cache.
//
// Parameters:
//
//	name: Name of the plugin to retrieve
//
// Returns the plugin (if found) and a boolean indicating existence.
func (l *PluginLoader) GetPlugin(name string) (DatabasePlugin, bool) {
	if name == "" {
		return nil, false
	}

	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	plugin, exists := l.cache[name]
	return plugin, exists
}

// UnloadPlugin unloads a plugin from memory and kills its process.
// This method is safe for concurrent use.
//
// Parameters:
//
//	name: Name of the plugin to unload
func (l *PluginLoader) UnloadPlugin(name string) {
	if name == "" {
		return
	}

	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	// Remove from cache
	delete(l.cache, name)

	// Kill the client process if it exists
	if client, exists := l.clients[name]; exists {
		client.Kill()
		delete(l.clients, name)
		// Unloaded plugin
	}
}

// ValidatePlugin validates that a plugin correctly implements the DatabasePlugin interface
// and has valid metadata. This includes checking that required methods return valid values.
//
// Parameters:
//
//	plugin: Plugin to validate
//
// Returns an error if validation fails, nil if the plugin is valid.
func (l *PluginLoader) ValidatePlugin(plugin DatabasePlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	// Validate plugin name
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	// Validate plugin version
	version := plugin.Version()
	if version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	// Validate supported engines
	engines := plugin.SupportedEngines()
	if len(engines) == 0 {
		return fmt.Errorf("plugin must support at least one database engine")
	}

	// Check for empty engine names
	for i, engine := range engines {
		if engine == "" {
			return fmt.Errorf("engine name at index %d cannot be empty", i)
		}
	}

	// Plugin validation successful

	return nil
}
