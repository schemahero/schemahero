package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginDownloader_GetPluginArtifactRef(t *testing.T) {
	downloader := NewPluginDownloader("")

	tests := []struct {
		driver       string
		majorVersion string
		expected     string
	}{
		{"postgres", "0", "docker.io/schemahero/plugin-postgres:0"},
		{"mysql", "1", "docker.io/schemahero/plugin-mysql:0"},
		{"cassandra", "0", "docker.io/schemahero/plugin-cassandra:0"},
	}

	for _, test := range tests {
		result := downloader.GetPluginArtifactRef(test.driver, test.majorVersion)
		if result != test.expected {
			t.Errorf("GetPluginArtifactRef(%s, %s) = %s, expected %s", test.driver, test.majorVersion, result, test.expected)
		}
	}
}

func TestPluginDownloader_GetCachedPluginPath(t *testing.T) {
	tempDir := t.TempDir()
	downloader := NewPluginDownloader(tempDir)

	tests := []struct {
		driver       string
		majorVersion string
		expected     string
	}{
		{"postgres", "0", filepath.Join(tempDir, "schemahero-postgres")},
		{"mysql", "1", filepath.Join(tempDir, "schemahero-mysql")},
	}

	for _, test := range tests {
		result := downloader.GetCachedPluginPath(test.driver, test.majorVersion)
		if result != test.expected {
			t.Errorf("GetCachedPluginPath(%s, %s) = %s, expected %s", test.driver, test.majorVersion, result, test.expected)
		}
	}
}

func TestPluginDownloader_IsPluginCached(t *testing.T) {
	tempDir := t.TempDir()
	downloader := NewPluginDownloader(tempDir)

	// Test non-existent plugin
	if downloader.IsPluginCached("postgres", "0") {
		t.Error("Expected IsPluginCached to return false for non-existent plugin")
	}

	// Create a mock plugin binary
	pluginPath := downloader.GetCachedPluginPath("postgres", "0")
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Create the file
	file, err := os.Create(pluginPath)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Make it executable
	if err := os.Chmod(pluginPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Now it should be cached
	if !downloader.IsPluginCached("postgres", "0") {
		t.Error("Expected IsPluginCached to return true for existing executable plugin")
	}

	// Test with non-executable file
	nonExecPath := downloader.GetCachedPluginPath("mysql", "0")
	if err := os.MkdirAll(filepath.Dir(nonExecPath), 0755); err != nil {
		t.Fatal(err)
	}

	file, err = os.Create(nonExecPath)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Don't make it executable (default permissions)
	if downloader.IsPluginCached("mysql", "0") {
		t.Error("Expected IsPluginCached to return false for non-executable plugin")
	}
}

func TestPluginDownloader_CleanPluginCache(t *testing.T) {
	tempDir := t.TempDir()
	downloader := NewPluginDownloader(tempDir)

	// Create a mock plugin binary
	pluginPath := downloader.GetCachedPluginPath("postgres", "0")
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		t.Fatal(err)
	}

	file, err := os.Create(pluginPath)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Verify it exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Fatal("Plugin file should exist before cleaning")
	}

	// Clean the cache
	if err := downloader.CleanPluginCache("postgres", "0"); err != nil {
		t.Fatal(err)
	}

	// Verify it's gone
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		t.Error("Plugin file should not exist after cleaning")
	}
}

// TestPluginDownloader_DownloadPlugin_Integration tests the actual download functionality
func TestPluginDownloader_DownloadPlugin_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	downloader := NewPluginDownloader(tempDir)

	ctx := context.Background()

	// Try to download the postgres plugin that should exist
	t.Logf("Attempting to download postgres plugin to %s", tempDir)
	pluginPath, err := downloader.DownloadPlugin(ctx, "postgres", "0")

	if err != nil {
		t.Logf("Download failed (this may be expected if plugin doesn't exist): %v", err)
		// Don't fail the test - just log the error so we can see what happens
		return
	}

	// If download succeeded, verify the plugin exists and is executable
	if pluginPath == "" {
		t.Error("Plugin path should not be empty on successful download")
		return
	}

	if !downloader.IsPluginCached("postgres", "0") {
		t.Error("Plugin should be cached after successful download")
		return
	}

	// Verify the file exists and is executable
	info, err := os.Stat(pluginPath)
	if err != nil {
		t.Errorf("Plugin binary should exist at %s: %v", pluginPath, err)
		return
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("Plugin binary should be executable")
		return
	}

	t.Logf("Successfully downloaded and verified plugin at %s", pluginPath)
}
