// Package plugin provides the interface and types for SchemaHero database plugins.
package plugin

import (
	"sync"
)

var (
	// globalPluginManager holds the global plugin manager instance
	globalPluginManager *PluginManager

	// globalPluginManagerMutex protects access to the global plugin manager
	globalPluginManagerMutex sync.RWMutex
)

// InitializePluginSystem initializes the global plugin system.
// This function should be called once during application startup
// to set up the plugin infrastructure.
//
// The function creates a new plugin registry, loader, and manager, then stores
// the manager instance globally for use throughout the application. The plugin
// loader is configured with the default plugin directory "/var/lib/schemahero/plugins".
//
// This function is thread-safe and can be called multiple times safely - subsequent
// calls will replace the existing plugin manager instance.
//
// Returns the initialized plugin manager instance.
func InitializePluginSystem() *PluginManager {
	globalPluginManagerMutex.Lock()
	defer globalPluginManagerMutex.Unlock()

	// Create plugin system components
	registry := NewPluginRegistry()
	loader := NewPluginLoader("/var/lib/schemahero/plugins")
	manager := NewPluginManager(registry, loader)

	// Set up bidirectional relationship between registry and loader
	registry.SetLoader(loader)

	// Store the global instance
	globalPluginManager = manager

	// Auto-discover plugins from common locations
	manager.DiscoverPlugins()

	return manager
}

// GetGlobalPluginManager returns the global plugin manager instance.
// This function provides thread-safe access to the globally initialized
// plugin manager for use throughout the application.
//
// If the plugin system has not been initialized via InitializePluginSystem,
// this function returns nil. Callers should check for nil return values
// and handle the case where the plugin system is not initialized.
//
// Returns the global plugin manager instance, or nil if not initialized.
func GetGlobalPluginManager() *PluginManager {
	globalPluginManagerMutex.RLock()
	defer globalPluginManagerMutex.RUnlock()

	return globalPluginManager
}

// ResetGlobalPluginSystem resets the global plugin system state.
// This function is primarily intended for testing purposes to ensure
// clean state between test cases.
//
// It sets the global plugin manager to nil.
// This function is thread-safe.
func ResetGlobalPluginSystem() {
	globalPluginManagerMutex.Lock()
	globalPluginManager = nil
	globalPluginManagerMutex.Unlock()
}
