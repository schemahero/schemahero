package plugin

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewPluginRegistry(t *testing.T) {
	registry := NewPluginRegistry()

	if registry == nil {
		t.Fatal("NewPluginRegistry() returned nil")
	}

	if registry.plugins == nil {
		t.Error("plugins map should be initialized")
	}

	if len(registry.plugins) != 0 {
		t.Error("plugins map should be empty initially")
	}

	defaultRepo := registry.GetDefaultRepo()
	if defaultRepo != "docker.io/schemahero/plugins" {
		t.Errorf("expected default repo 'docker.io/schemahero/plugins', got '%s'", defaultRepo)
	}

	if registry.GetLoader() != nil {
		t.Error("loader should be nil initially")
	}
}

func TestRegisterPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &PluginInfo{
		Name:      "postgres",
		Version:   "v1.0.0",
		OCIRef:    "docker.io/schemahero/plugins/postgres:v1.0.0",
		Digest:    "sha256:1234567890abcdef",
		Platform:  "linux/amd64",
		Engines:   []string{"postgres"},
		LocalPath: "/path/to/postgres-plugin",
	}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Verify plugin was registered
	retrievedPlugin, err := registry.Get("postgres")
	if err != nil {
		t.Fatalf("Get() failed after registration: %v", err)
	}

	if retrievedPlugin.Name != plugin.Name {
		t.Errorf("expected name '%s', got '%s'", plugin.Name, retrievedPlugin.Name)
	}

	if retrievedPlugin.Version != plugin.Version {
		t.Errorf("expected version '%s', got '%s'", plugin.Version, retrievedPlugin.Version)
	}

	if retrievedPlugin.OCIRef != plugin.OCIRef {
		t.Errorf("expected OCIRef '%s', got '%s'", plugin.OCIRef, retrievedPlugin.OCIRef)
	}

	if len(retrievedPlugin.Engines) != len(plugin.Engines) {
		t.Errorf("expected %d engines, got %d", len(plugin.Engines), len(retrievedPlugin.Engines))
	}
}

func TestRegisterDuplicatePlugin(t *testing.T) {
	registry := NewPluginRegistry()

	plugin1 := &PluginInfo{
		Name:    "postgres",
		Version: "v1.0.0",
		Engines: []string{"postgres"},
	}

	plugin2 := &PluginInfo{
		Name:    "postgres", // Same name
		Version: "v2.0.0",   // Different version
		Engines: []string{"postgres"},
	}

	// Register first plugin
	err := registry.Register(plugin1)
	if err != nil {
		t.Fatalf("Register() failed for first plugin: %v", err)
	}

	// Register second plugin with same name (should replace)
	err = registry.Register(plugin2)
	if err != nil {
		t.Fatalf("Register() failed for duplicate plugin: %v", err)
	}

	// Verify the second plugin replaced the first
	retrievedPlugin, err := registry.Get("postgres")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrievedPlugin.Version != plugin2.Version {
		t.Errorf("expected version '%s', got '%s'", plugin2.Version, retrievedPlugin.Version)
	}

	// Verify only one plugin exists
	plugins := registry.List()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
}

func TestRegisterInvalidPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	tests := []struct {
		name     string
		plugin   *PluginInfo
		expected string
	}{
		{
			name:     "nil plugin",
			plugin:   nil,
			expected: "plugin info cannot be nil",
		},
		{
			name: "empty name",
			plugin: &PluginInfo{
				Name:    "",
				Version: "v1.0.0",
				Engines: []string{"postgres"},
			},
			expected: "plugin name cannot be empty",
		},
		{
			name: "empty version",
			plugin: &PluginInfo{
				Name:    "postgres",
				Version: "",
				Engines: []string{"postgres"},
			},
			expected: "plugin version cannot be empty",
		},
		{
			name: "no engines",
			plugin: &PluginInfo{
				Name:    "postgres",
				Version: "v1.0.0",
				Engines: []string{},
			},
			expected: "plugin must support at least one database engine",
		},
		{
			name: "nil engines",
			plugin: &PluginInfo{
				Name:    "postgres",
				Version: "v1.0.0",
				Engines: nil,
			},
			expected: "plugin must support at least one database engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.plugin)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if err.Error() != tt.expected {
				t.Errorf("expected error '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestGetPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &PluginInfo{
		Name:      "mysql",
		Version:   "v1.5.0",
		OCIRef:    "docker.io/schemahero/plugins/mysql:v1.5.0",
		Digest:    "sha256:abcdef1234567890",
		Platform:  "linux/amd64",
		Engines:   []string{"mysql", "mariadb"},
		LocalPath: "/path/to/mysql-plugin",
	}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	retrievedPlugin, err := registry.Get("mysql")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Verify all fields
	if retrievedPlugin.Name != plugin.Name {
		t.Errorf("expected name '%s', got '%s'", plugin.Name, retrievedPlugin.Name)
	}
	if retrievedPlugin.Version != plugin.Version {
		t.Errorf("expected version '%s', got '%s'", plugin.Version, retrievedPlugin.Version)
	}
	if retrievedPlugin.OCIRef != plugin.OCIRef {
		t.Errorf("expected OCIRef '%s', got '%s'", plugin.OCIRef, retrievedPlugin.OCIRef)
	}
	if retrievedPlugin.Digest != plugin.Digest {
		t.Errorf("expected digest '%s', got '%s'", plugin.Digest, retrievedPlugin.Digest)
	}
	if retrievedPlugin.Platform != plugin.Platform {
		t.Errorf("expected platform '%s', got '%s'", plugin.Platform, retrievedPlugin.Platform)
	}
	if retrievedPlugin.LocalPath != plugin.LocalPath {
		t.Errorf("expected LocalPath '%s', got '%s'", plugin.LocalPath, retrievedPlugin.LocalPath)
	}

	// Verify engines slice is copied correctly
	if len(retrievedPlugin.Engines) != len(plugin.Engines) {
		t.Errorf("expected %d engines, got %d", len(plugin.Engines), len(retrievedPlugin.Engines))
	}
	for i, engine := range plugin.Engines {
		if i >= len(retrievedPlugin.Engines) || retrievedPlugin.Engines[i] != engine {
			t.Errorf("expected engine[%d] '%s', got '%s'", i, engine, retrievedPlugin.Engines[i])
		}
	}

	// Test that modifying retrieved plugin doesn't affect original
	retrievedPlugin.Version = "modified"
	retrievedPlugin.Engines[0] = "modified"

	secondRetrieval, err := registry.Get("mysql")
	if err != nil {
		t.Fatalf("second Get() failed: %v", err)
	}
	if secondRetrieval.Version != plugin.Version {
		t.Error("modifying retrieved plugin affected original")
	}
	if secondRetrieval.Engines[0] != plugin.Engines[0] {
		t.Error("modifying retrieved plugin engines affected original")
	}
}

func TestGetNonExistentPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	tests := []struct {
		name        string
		pluginName  string
		expectedErr string
	}{
		{
			name:        "empty name",
			pluginName:  "",
			expectedErr: "plugin name cannot be empty",
		},
		{
			name:        "non-existent plugin",
			pluginName:  "nonexistent",
			expectedErr: "plugin 'nonexistent' not found in registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := registry.Get(tt.pluginName)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if plugin != nil {
				t.Error("expected nil plugin when error occurs")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestListPlugins(t *testing.T) {
	registry := NewPluginRegistry()

	// Test empty registry
	plugins := registry.List()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins in empty registry, got %d", len(plugins))
	}

	// Add multiple plugins
	plugin1 := &PluginInfo{
		Name:    "postgres",
		Version: "v1.0.0",
		Engines: []string{"postgres"},
	}
	plugin2 := &PluginInfo{
		Name:    "mysql",
		Version: "v1.5.0",
		Engines: []string{"mysql"},
	}
	plugin3 := &PluginInfo{
		Name:    "sqlite",
		Version: "v2.0.0",
		Engines: []string{"sqlite"},
	}

	for _, plugin := range []*PluginInfo{plugin1, plugin2, plugin3} {
		err := registry.Register(plugin)
		if err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	// Test listing all plugins
	plugins = registry.List()
	if len(plugins) != 3 {
		t.Errorf("expected 3 plugins, got %d", len(plugins))
	}

	// Verify all plugins are present (order may vary)
	names := make(map[string]bool)
	for _, plugin := range plugins {
		names[plugin.Name] = true
	}

	expectedNames := []string{"postgres", "mysql", "sqlite"}
	for _, expected := range expectedNames {
		if !names[expected] {
			t.Errorf("expected plugin '%s' not found in list", expected)
		}
	}

	// Test that modifying returned plugins doesn't affect registry
	plugins[0].Version = "modified"
	plugins[0].Engines[0] = "modified"

	secondList := registry.List()
	found := false
	for _, plugin := range secondList {
		if plugin.Name == plugins[0].Name {
			found = true
			if plugin.Version == "modified" {
				t.Error("modifying listed plugin affected registry")
			}
			break
		}
	}
	if !found {
		t.Error("could not find plugin in second list")
	}
}

func TestRemovePlugin(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &PluginInfo{
		Name:    "postgres",
		Version: "v1.0.0",
		Engines: []string{"postgres"},
	}

	// Register plugin
	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Verify it exists
	_, err = registry.Get("postgres")
	if err != nil {
		t.Fatalf("Get() failed before removal: %v", err)
	}

	// Remove plugin
	err = registry.Remove("postgres")
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify it no longer exists
	_, err = registry.Get("postgres")
	if err == nil {
		t.Error("Get() should have failed after removal")
	}

	// Verify registry is empty
	plugins := registry.List()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins after removal, got %d", len(plugins))
	}
}

func TestRemoveNonExistentPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	tests := []struct {
		name        string
		pluginName  string
		expectedErr string
	}{
		{
			name:        "empty name",
			pluginName:  "",
			expectedErr: "plugin name cannot be empty",
		},
		{
			name:        "non-existent plugin",
			pluginName:  "nonexistent",
			expectedErr: "plugin 'nonexistent' not found in registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Remove(tt.pluginName)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestThreadSafety(t *testing.T) {
	registry := NewPluginRegistry()
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup

	// Test concurrent registration
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				plugin := &PluginInfo{
					Name:    fmt.Sprintf("plugin-%d-%d", id, j),
					Version: "v1.0.0",
					Engines: []string{"test"},
				}
				err := registry.Register(plugin)
				if err != nil {
					t.Errorf("Register() failed in goroutine %d: %v", id, err)
				}
			}
		}(i)
	}
	wg.Wait()

	// Verify all plugins were registered
	plugins := registry.List()
	expectedCount := numGoroutines * numOperations
	if len(plugins) != expectedCount {
		t.Errorf("expected %d plugins, got %d", expectedCount, len(plugins))
	}

	// Test concurrent read/write operations
	wg.Add(numGoroutines * 3) // readers, writers, removers

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				name := fmt.Sprintf("plugin-%d-%d", id%numGoroutines, j%numOperations)
				_, err := registry.Get(name)
				// Error is okay since removers might delete plugins
				_ = err
			}
		}(i)
	}

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				plugin := &PluginInfo{
					Name:    fmt.Sprintf("new-plugin-%d-%d", id, j),
					Version: "v2.0.0",
					Engines: []string{"test"},
				}
				err := registry.Register(plugin)
				if err != nil {
					t.Errorf("Register() failed in writer goroutine %d: %v", id, err)
				}
			}
		}(i)
	}

	// Removers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				name := fmt.Sprintf("plugin-%d-%d", id%numGoroutines, j%numOperations)
				err := registry.Remove(name)
				// Error is okay since multiple goroutines might try to remove the same plugin
				_ = err
			}
		}(i)
	}

	wg.Wait()

	// Test concurrent List operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				plugins := registry.List()
				// Just ensure it doesn't panic or race
				_ = plugins
			}
		}()
	}
	wg.Wait()
}

func TestSetGetLoader(t *testing.T) {
	registry := NewPluginRegistry()

	// Test initial state
	if registry.GetLoader() != nil {
		t.Error("loader should be nil initially")
	}

	// Create a loader
	loader := NewPluginLoader("/test/path")
	if loader == nil {
		t.Fatal("NewPluginLoader() returned nil")
	}

	// Set loader
	registry.SetLoader(loader)

	// Get loader and verify
	retrievedLoader := registry.GetLoader()
	if retrievedLoader != loader {
		t.Error("GetLoader() returned different loader instance")
	}

	// Test setting nil loader
	registry.SetLoader(nil)
	if registry.GetLoader() != nil {
		t.Error("loader should be nil after setting to nil")
	}
}

func TestDefaultRepo(t *testing.T) {
	registry := NewPluginRegistry()

	// Test initial default repo
	defaultRepo := registry.GetDefaultRepo()
	if defaultRepo != "docker.io/schemahero/plugins" {
		t.Errorf("expected default repo 'docker.io/schemahero/plugins', got '%s'", defaultRepo)
	}

	// Test setting custom repo
	customRepo := "custom.registry.com/my-plugins"
	registry.SetDefaultRepo(customRepo)

	retrievedRepo := registry.GetDefaultRepo()
	if retrievedRepo != customRepo {
		t.Errorf("expected repo '%s', got '%s'", customRepo, retrievedRepo)
	}

	// Test setting empty repo
	registry.SetDefaultRepo("")
	retrievedRepo = registry.GetDefaultRepo()
	if retrievedRepo != "" {
		t.Errorf("expected empty repo, got '%s'", retrievedRepo)
	}

	// Test thread safety of repo operations
	var wg sync.WaitGroup
	const numGoroutines = 10

	wg.Add(numGoroutines * 2)

	// Setters
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			repo := fmt.Sprintf("repo-%d.com/plugins", id)
			registry.SetDefaultRepo(repo)
		}(i)
	}

	// Getters
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			repo := registry.GetDefaultRepo()
			// Just ensure it doesn't panic or race
			_ = repo
		}()
	}

	wg.Wait()
}

// Helper function to compare PluginInfo structs
func comparePluginInfo(t *testing.T, expected, actual *PluginInfo) {
	t.Helper()

	if actual.Name != expected.Name {
		t.Errorf("expected name '%s', got '%s'", expected.Name, actual.Name)
	}
	if actual.Version != expected.Version {
		t.Errorf("expected version '%s', got '%s'", expected.Version, actual.Version)
	}
	if actual.OCIRef != expected.OCIRef {
		t.Errorf("expected OCIRef '%s', got '%s'", expected.OCIRef, actual.OCIRef)
	}
	if actual.Digest != expected.Digest {
		t.Errorf("expected digest '%s', got '%s'", expected.Digest, actual.Digest)
	}
	if actual.Platform != expected.Platform {
		t.Errorf("expected platform '%s', got '%s'", expected.Platform, actual.Platform)
	}
	if actual.LocalPath != expected.LocalPath {
		t.Errorf("expected LocalPath '%s', got '%s'", expected.LocalPath, actual.LocalPath)
	}
	if len(actual.Engines) != len(expected.Engines) {
		t.Errorf("expected %d engines, got %d", len(expected.Engines), len(actual.Engines))
		return
	}
	for i, engine := range expected.Engines {
		if actual.Engines[i] != engine {
			t.Errorf("expected engine[%d] '%s', got '%s'", i, engine, actual.Engines[i])
		}
	}
}
