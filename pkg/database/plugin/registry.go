// Package plugin provides the interface and types for SchemaHero database plugins.
package plugin

import (
	"fmt"
	"sync"
)

// PluginRegistry manages the registration and retrieval of database plugins.
// It provides thread-safe operations for managing plugin information and
// maintains a registry of available plugins that can be loaded and used
// by SchemaHero.
type PluginRegistry struct {
	mu          sync.RWMutex
	plugins     map[string]*PluginInfo
	loader      *PluginLoader
	defaultRepo string
}

// PluginInfo contains metadata about a registered plugin, including
// its location, version, and supported database engines.
type PluginInfo struct {
	// Name is the unique identifier for the plugin
	Name string

	// Version is the semantic version of the plugin
	Version string

	// OCIRef is the full OCI reference for the plugin
	// (e.g., "docker.io/schemahero/plugins/postgres:v1.2.3")
	OCIRef string

	// Digest is the OCI content digest for content verification
	Digest string

	// Platform specifies the target platform (linux/amd64, darwin/arm64, etc.)
	Platform string

	// Engines lists the database engines supported by this plugin
	Engines []string

	// LocalPath is the file system path to the cached plugin binary
	LocalPath string
}

// PluginSource represents the source configuration for a plugin,
// supporting both OCI registry and local file sources.
type PluginSource struct {
	// Type specifies the source type ("oci" or "local")
	Type string

	// Registry is the OCI registry URL (used for "oci" type)
	Registry string

	// Repo is the repository path within the registry (used for "oci" type)
	Repo string

	// Tag is the version tag (used for "oci" type)
	Tag string

	// Path is the local file system path (used for "local" type)
	Path string
}

// NewPluginRegistry creates a new plugin registry with default configuration.
// The registry is initialized with an empty plugin map and the default
// OCI repository set to "docker.io/schemahero/plugins".
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:     make(map[string]*PluginInfo),
		defaultRepo: "docker.io/schemahero/plugins",
	}
}

// Register adds a plugin to the registry. If a plugin with the same name
// already exists, it will be replaced with the new information.
//
// Parameters:
//
//	info: Plugin information to register
//
// Returns an error if the plugin info is invalid.
func (r *PluginRegistry) Register(info *PluginInfo) error {
	if info == nil {
		return fmt.Errorf("plugin info cannot be nil")
	}

	if info.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if info.Version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	if len(info.Engines) == 0 {
		return fmt.Errorf("plugin must support at least one database engine")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins[info.Name] = info
	return nil
}

// Get retrieves plugin information by name. Returns the plugin info
// if found, or an error if the plugin is not registered.
//
// Parameters:
//
//	name: The name of the plugin to retrieve
//
// Returns the plugin info or an error if not found.
func (r *PluginRegistry) Get(name string) (*PluginInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found in registry", name)
	}

	// Return a copy to prevent external modification
	return &PluginInfo{
		Name:      info.Name,
		Version:   info.Version,
		OCIRef:    info.OCIRef,
		Digest:    info.Digest,
		Platform:  info.Platform,
		Engines:   append([]string(nil), info.Engines...),
		LocalPath: info.LocalPath,
	}, nil
}

// List returns a slice of all registered plugins. The returned slice
// contains copies of the plugin information to prevent external modification.
//
// Returns a slice of all plugin information in the registry.
func (r *PluginRegistry) List() []*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*PluginInfo, 0, len(r.plugins))
	for _, info := range r.plugins {
		// Create a copy to prevent external modification
		plugins = append(plugins, &PluginInfo{
			Name:      info.Name,
			Version:   info.Version,
			OCIRef:    info.OCIRef,
			Digest:    info.Digest,
			Platform:  info.Platform,
			Engines:   append([]string(nil), info.Engines...),
			LocalPath: info.LocalPath,
		})
	}

	return plugins
}

// Remove removes a plugin from the registry by name.
//
// Parameters:
//
//	name: The name of the plugin to remove
//
// Returns an error if the plugin name is empty or if the plugin is not found.
func (r *PluginRegistry) Remove(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found in registry", name)
	}

	delete(r.plugins, name)
	return nil
}

// SetLoader sets the plugin loader for this registry. The loader is responsible
// for downloading, caching, and loading plugin binaries.
//
// Parameters:
//
//	loader: The plugin loader to associate with this registry
func (r *PluginRegistry) SetLoader(loader *PluginLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loader = loader
}

// GetLoader returns the current plugin loader associated with this registry.
// Returns nil if no loader has been set.
func (r *PluginRegistry) GetLoader() *PluginLoader {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loader
}

// GetDefaultRepo returns the default OCI repository URL used for plugin resolution.
func (r *PluginRegistry) GetDefaultRepo() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultRepo
}

// SetDefaultRepo sets the default OCI repository URL for plugin resolution.
//
// Parameters:
//
//	repo: The default repository URL to use
func (r *PluginRegistry) SetDefaultRepo(repo string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultRepo = repo
}
