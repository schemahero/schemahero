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
		{"postgres", "0", "schemahero/plugin-postgres:0"},
		{"mysql", "1", "schemahero/plugin-mysql:1"},
		{"cassandra", "0", "schemahero/plugin-cassandra:0"},
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

// TestPluginDownloader_DownloadPlugin tests the download functionality with a mock
// Note: This test will fail in real execution since we don't have actual ORAS artifacts
// It's mainly for demonstrating the API and ensuring compilation
func TestPluginDownloader_DownloadPlugin(t *testing.T) {
	t.Skip("Skipping download test - requires actual ORAS artifacts")

	tempDir := t.TempDir()
	downloader := NewPluginDownloader(tempDir)

	ctx := context.Background()

	// This would attempt to download schemahero/plugin-postgres:0
	_, err := downloader.DownloadPlugin(ctx, "postgres", "0")

	// We expect this to fail since the artifacts don't exist yet
	if err == nil {
		t.Error("Expected download to fail for non-existent artifact")
	}
}
