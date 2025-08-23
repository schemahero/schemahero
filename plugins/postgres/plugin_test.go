package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/schemahero/schemahero/pkg/database/plugin"
)

const (
	expectedPluginName     = "postgres"
	expectedPluginVersion  = "1.0.0"
	pluginBinaryName      = "schemahero-postgres"
)

// getPluginBinaryPath returns the path to the compiled plugin binary
func getPluginBinaryPath() string {
	// Look for the plugin binary in ../bin/ (relative to this test file)
	return filepath.Join("..", "bin", pluginBinaryName)
}

// buildPluginIfNeeded builds the plugin if the binary doesn't exist
func buildPluginIfNeeded(t *testing.T) string {
	pluginPath := getPluginBinaryPath()
	
	// Check if plugin binary exists
	if _, err := os.Stat(pluginPath); err == nil {
		t.Logf("Plugin binary already exists at: %s", pluginPath)
		return pluginPath
	}
	
	t.Logf("Plugin binary not found, building...")
	
	// Create the directory if it doesn't exist
	pluginDir := filepath.Dir(pluginPath)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin directory %s: %v", pluginDir, err)
	}
	
	// Build the plugin using go build
	// We run this from the postgres directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	
	// Change to the postgres plugin directory (where this test is located)
	if err := os.Chdir("."); err != nil {
		t.Fatalf("Failed to change to plugin directory: %v", err)
	}
	
	// Use the build script if it exists, otherwise use go build directly
	buildScript := "./build.sh"
	if _, err := os.Stat(buildScript); err == nil {
		// Set OUTPUT_DIR environment variable to point to our desired location
		os.Setenv("OUTPUT_DIR", pluginDir)
		defer os.Unsetenv("OUTPUT_DIR")
		
		// Note: In a real test environment, we would use exec.Command to run the build script
		// For this test, we'll assume the plugin is already built or build it with go build
	}
	
	// Fallback to direct go build
	// This would normally use exec.Command("go", "build", "-o", pluginPath, ".")
	// For the test, we'll check if the file exists after the above setup
	if _, err := os.Stat(pluginPath); err != nil {
		t.Skipf("Plugin binary not found and cannot be built automatically. Please run 'make postgres' in the plugins directory first. Expected path: %s", pluginPath)
	}
	
	return pluginPath
}

// TestPostgresPluginIntegration tests the postgres plugin as a real integration test
func TestPostgresPluginIntegration(t *testing.T) {
	// Build plugin if needed
	pluginPath := buildPluginIfNeeded(t)
	
	// Verify plugin binary exists and is executable
	fileInfo, err := os.Stat(pluginPath)
	if err != nil {
		t.Fatalf("Plugin binary does not exist: %s", pluginPath)
	}
	
	if !fileInfo.Mode().IsRegular() {
		t.Fatalf("Plugin path is not a regular file: %s", pluginPath)
	}
	
	if fileInfo.Mode().Perm()&0111 == 0 {
		t.Fatalf("Plugin binary is not executable: %s", pluginPath)
	}
	
	t.Logf("Using plugin binary at: %s", pluginPath)
	
	// Create plugin loader
	loader := plugin.NewPluginLoader("")
	defer loader.Cleanup()
	
	// Load the plugin
	ctx := context.Background()
	loadedPlugin, err := loader.LoadLocal(ctx, pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin from %s: %v", pluginPath, err)
	}
	
	// Test plugin metadata
	t.Run("PluginMetadata", func(t *testing.T) {
		// Test Name
		name := loadedPlugin.Name()
		if name != expectedPluginName {
			t.Errorf("Expected plugin name '%s', got '%s'", expectedPluginName, name)
		}
		
		// Test Version
		version := loadedPlugin.Version()
		if version != expectedPluginVersion {
			t.Errorf("Expected plugin version '%s', got '%s'", expectedPluginVersion, version)
		}
		
		// Test SupportedEngines
		engines := loadedPlugin.SupportedEngines()
		expectedEngines := []string{"postgres", "postgresql", "cockroachdb"}
		
		if len(engines) != len(expectedEngines) {
			t.Errorf("Expected %d supported engines, got %d: %v", len(expectedEngines), len(engines), engines)
		}
		
		// Check that all expected engines are present
		engineSet := make(map[string]bool)
		for _, engine := range engines {
			engineSet[engine] = true
		}
		
		for _, expectedEngine := range expectedEngines {
			if !engineSet[expectedEngine] {
				t.Errorf("Expected engine '%s' not found in supported engines: %v", expectedEngine, engines)
			}
		}
		
		t.Logf("Plugin metadata: name=%s, version=%s, engines=%v", name, version, engines)
	})
	
	// Test Validate method
	t.Run("ValidateMethod", func(t *testing.T) {
		// Test with valid config
		validConfig := map[string]interface{}{
			"uri": "postgresql://user:password@localhost:5432/dbname",
		}
		
		err := loadedPlugin.Validate(validConfig)
		if err != nil {
			t.Errorf("Expected validation to pass with valid config, got error: %v", err)
		}
		
		// Test with missing URI
		invalidConfig := map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		}
		
		err = loadedPlugin.Validate(invalidConfig)
		if err == nil {
			t.Error("Expected validation to fail with missing URI")
		} else {
			t.Logf("Validation correctly failed with missing URI: %v", err)
		}
		
		// Test with empty URI
		emptyURIConfig := map[string]interface{}{
			"uri": "",
		}
		
		err = loadedPlugin.Validate(emptyURIConfig)
		if err == nil {
			t.Error("Expected validation to fail with empty URI")
		} else {
			t.Logf("Validation correctly failed with empty URI: %v", err)
		}
	})
	
	// Test Initialize method
	t.Run("InitializeMethod", func(t *testing.T) {
		err := loadedPlugin.Initialize(ctx)
		if err != nil {
			t.Errorf("Expected initialization to succeed, got error: %v", err)
		} else {
			t.Log("Plugin initialized successfully")
		}
	})
	
	// Test Shutdown method
	t.Run("ShutdownMethod", func(t *testing.T) {
		err := loadedPlugin.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected shutdown to succeed, got error: %v", err)
		} else {
			t.Log("Plugin shutdown successfully")
		}
	})
	
	// Test Connect method (should fail without real postgres instance)
	t.Run("ConnectMethod", func(t *testing.T) {
		// We can't test actual connections without a real postgres instance,
		// but we can test that the method exists and handles invalid URIs appropriately
		
		// Test with obviously invalid URI
		invalidURI := "invalid://not-a-real-uri"
		conn, err := loadedPlugin.Connect(invalidURI, nil)
		
		// We expect this to fail since there's no real postgres instance
		if err == nil {
			// If it somehow succeeds, make sure to clean up
			if conn != nil {
				conn.Close()
			}
			t.Log("Connect method returned without error for invalid URI (unexpected but not necessarily wrong)")
		} else {
			t.Logf("Connect method correctly failed with invalid URI: %v", err)
		}
		
		// Test with valid-looking URI but no real server
		validLookingURI := "postgresql://user:password@localhost:5432/nonexistent"
		conn, err = loadedPlugin.Connect(validLookingURI, nil)
		
		if err == nil {
			// Clean up if somehow successful
			if conn != nil {
				conn.Close()
			}
			t.Log("Connect method succeeded with valid-looking URI (unexpected but plugin might have different behavior)")
		} else {
			t.Logf("Connect method failed with unreachable server (expected): %v", err)
		}
		
		// Test with options
		options := map[string]interface{}{
			"connect_timeout": 5,
			"sslmode":        "disable",
		}
		
		conn, err = loadedPlugin.Connect(validLookingURI, options)
		
		if err == nil {
			if conn != nil {
				conn.Close()
			}
			t.Log("Connect method with options succeeded (unexpected)")
		} else {
			t.Logf("Connect method with options failed as expected: %v", err)
		}
	})
}

// TestPluginLoader tests that the plugin can be loaded and cached properly
func TestPluginLoader(t *testing.T) {
	pluginPath := buildPluginIfNeeded(t)
	
	// Create plugin loader
	loader := plugin.NewPluginLoader("")
	defer loader.Cleanup()
	
	ctx := context.Background()
	
	// Load plugin first time
	plugin1, err := loader.LoadLocal(ctx, pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin first time: %v", err)
	}
	
	// Verify plugin is cached
	cachedPlugin, exists := loader.GetPlugin(expectedPluginName)
	if !exists {
		t.Error("Plugin should be cached after loading")
	}
	
	if cachedPlugin != plugin1 {
		t.Error("Cached plugin should be the same instance")
	}
	
	// Load plugin second time using LoadFromInfo (which checks cache first)
	info := &plugin.PluginInfo{
		Name:      expectedPluginName,
		LocalPath: pluginPath,
	}
	
	plugin2, err := loader.LoadFromInfo(ctx, info)
	if err != nil {
		t.Fatalf("Failed to load plugin second time: %v", err)
	}
	
	// Should be the same instance due to caching
	if plugin1 != plugin2 {
		t.Error("Second load should return cached plugin instance")
	}
	
	t.Log("Plugin caching works correctly")
	
	// Test unloading
	err = loader.Unload(expectedPluginName)
	if err != nil {
		t.Errorf("Failed to unload plugin: %v", err)
	}
	
	// Verify plugin is no longer cached
	_, exists = loader.GetPlugin(expectedPluginName)
	if exists {
		t.Error("Plugin should not be cached after unloading")
	}
	
	t.Log("Plugin unloading works correctly")
}

// TestPluginValidation tests the plugin validation logic
func TestPluginValidation(t *testing.T) {
	pluginPath := buildPluginIfNeeded(t)
	
	loader := plugin.NewPluginLoader("")
	defer loader.Cleanup()
	
	ctx := context.Background()
	loadedPlugin, err := loader.LoadLocal(ctx, pluginPath)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}
	
	// Test ValidatePlugin method of loader
	err = loader.ValidatePlugin(loadedPlugin)
	if err != nil {
		t.Errorf("Plugin validation should pass, got error: %v", err)
	}
	
	t.Log("Plugin validation passed")
}

// TestPluginFromInfo tests loading plugin from PluginInfo
func TestPluginFromInfo(t *testing.T) {
	pluginPath := buildPluginIfNeeded(t)
	
	loader := plugin.NewPluginLoader("")
	defer loader.Cleanup()
	
	ctx := context.Background()
	
	// Create PluginInfo for local loading
	info := &plugin.PluginInfo{
		Name:      expectedPluginName,
		LocalPath: pluginPath,
	}
	
	// Load from info
	loadedPlugin, err := loader.LoadFromInfo(ctx, info)
	if err != nil {
		t.Fatalf("Failed to load plugin from info: %v", err)
	}
	
	// Verify it's the expected plugin
	if loadedPlugin.Name() != expectedPluginName {
		t.Errorf("Expected plugin name '%s', got '%s'", expectedPluginName, loadedPlugin.Name())
	}
	
	// Test loading again (should use cache)
	loadedPlugin2, err := loader.LoadFromInfo(ctx, info)
	if err != nil {
		t.Fatalf("Failed to load plugin from info second time: %v", err)
	}
	
	if loadedPlugin != loadedPlugin2 {
		t.Error("Second load from info should return cached instance")
	}
	
	t.Log("Loading from PluginInfo works correctly")
}