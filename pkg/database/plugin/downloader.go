package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"  
	"oras.land/oras-go/v2/registry/remote"
)

// PluginDownloader handles downloading plugins from ORAS artifacts on Docker Hub
type PluginDownloader struct {
	cacheDir     string
	downloadOnce map[string]*sync.Once
	onceMutex    sync.RWMutex
}

// pluginRegistryOverride can be set at build time to override the default plugin registry
var pluginRegistryOverride string
var pluginTagOverride string

// SetPluginRegistryOverride allows runtime override of the plugin registry
func SetPluginRegistryOverride(registry string) {
	pluginRegistryOverride = registry
}

// SetPluginTagOverride allows runtime override of the plugin tag
func SetPluginTagOverride(tag string) {
	pluginTagOverride = tag
}

// NewPluginDownloader creates a new plugin downloader with specified cache directory
func NewPluginDownloader(cacheDir string) *PluginDownloader {
	if cacheDir == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			cacheDir = filepath.Join(homeDir, ".schemahero", "plugins")
		} else {
			cacheDir = "/tmp/plugins"
		}
	}

	return &PluginDownloader{
		cacheDir:     cacheDir,
		downloadOnce: make(map[string]*sync.Once),
		onceMutex:    sync.RWMutex{},
	}
}

// GetPluginArtifactRef returns the ORAS artifact reference for a given driver
// Maps driver names to Docker Hub artifact references following pattern:
// schemahero/plugin-{driver}:{major-version}
// Can be overridden at build time via pluginRegistryOverride
func (d *PluginDownloader) GetPluginArtifactRef(driver string, majorVersion string) string {
	if pluginRegistryOverride != "" {
		// For dev builds, append architecture suffix to get the right binary
		arch := runtime.GOARCH
		ref := fmt.Sprintf("%s-%s:%s-%s", pluginRegistryOverride, driver, majorVersion, arch)
		return ref
	}
	return fmt.Sprintf("schemahero/plugin-%s:%s", driver, majorVersion)
}

// GetCachedPluginPath returns the expected path for a cached plugin binary
func (d *PluginDownloader) GetCachedPluginPath(driver string, majorVersion string) string {
	pluginBinary := fmt.Sprintf("schemahero-%s", driver)
	return filepath.Join(d.cacheDir, pluginBinary)
}

// IsPluginCached checks if a plugin is already downloaded and cached locally
func (d *PluginDownloader) IsPluginCached(driver string, majorVersion string) bool {
	pluginPath := d.GetCachedPluginPath(driver, majorVersion)

	// Check if file exists and is executable
	if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
		// Check if file is executable
		return info.Mode().Perm()&0111 != 0
	}

	return false
}

// DownloadPlugin downloads a plugin from ORAS artifact if not already cached
// Returns the path to the plugin binary
func (d *PluginDownloader) DownloadPlugin(ctx context.Context, driver string, majorVersion string) (string, error) {
	if driver == "" {
		return "", fmt.Errorf("driver cannot be empty")
	}

	if majorVersion == "" {
		return "", fmt.Errorf("major version cannot be empty")
	}

	pluginPath := d.GetCachedPluginPath(driver, majorVersion)

	// Check if already cached
	if d.IsPluginCached(driver, majorVersion) {
		return pluginPath, nil
	}

	// Use sync.Once to ensure only one download per plugin happens concurrently
	downloadKey := fmt.Sprintf("%s:%s", driver, majorVersion)

	d.onceMutex.Lock()
	once, exists := d.downloadOnce[downloadKey]
	if !exists {
		once = &sync.Once{}
		d.downloadOnce[downloadKey] = once
	}
	d.onceMutex.Unlock()

	var downloadErr error
	once.Do(func() {
		downloadErr = d.downloadPluginOnce(ctx, driver, majorVersion, pluginPath)
	})

	if downloadErr != nil {
		// Clean up the sync.Once entry on failure so we can retry later
		d.onceMutex.Lock()
		delete(d.downloadOnce, downloadKey)
		d.onceMutex.Unlock()
		return "", downloadErr
	}

	return pluginPath, nil
}

// downloadPluginOnce performs the actual download operation
func (d *PluginDownloader) downloadPluginOnce(ctx context.Context, driver string, majorVersion string, pluginPath string) error {
	artifactRef := d.GetPluginArtifactRef(driver, majorVersion)
	
	// Extract the tag from the artifact reference (everything after the last :)
	parts := strings.Split(artifactRef, ":")
	tag := parts[len(parts)-1]

	// Ensure cache directory exists
	cacheDir := filepath.Dir(pluginPath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	// Create repository reference
	repo, err := remote.NewRepository(artifactRef)
	if err != nil {
		return fmt.Errorf("failed to create repository reference for %s: %w", artifactRef, err)
	}

	// Create file store and copy to it
	fs, err := file.New(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to download plugin %s: %w", artifactRef, err)
	}
	
	// ORAS seems to preserve directory structure, so look in common subdirectories
	possiblePaths := []string{
		filepath.Join(cacheDir, fmt.Sprintf("schemahero-%s", driver)),           // Direct
		filepath.Join(cacheDir, "plugins", "bin", fmt.Sprintf("schemahero-%s", driver)), // With structure
		filepath.Join(cacheDir, "plugins", fmt.Sprintf("schemahero-%s", driver)),        // Partial structure
	}
	
	var foundPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			fmt.Printf("[DEBUG] Found plugin at: %s\n", path)
			break
		}
	}
	
	if foundPath == "" {
		return fmt.Errorf("plugin not found in any expected location after download")
	}
	
	expectedFile := foundPath

	// Move to final location if needed
	if expectedFile != pluginPath {
		if err := os.Rename(expectedFile, pluginPath); err != nil {
			return fmt.Errorf("failed to move plugin: %w", err)
		}
	}

	// Ensure the binary is executable
	if err := os.Chmod(pluginPath, 0755); err != nil {
		return fmt.Errorf("failed to make plugin binary executable: %w", err)
	}

	return nil
}

// CleanCache removes all cached plugins
func (d *PluginDownloader) CleanCache() error {
	return os.RemoveAll(d.cacheDir)
}

// CleanPluginCache removes cached files for a specific plugin version
func (d *PluginDownloader) CleanPluginCache(driver string, majorVersion string) error {
	pluginDir := filepath.Join(d.cacheDir, majorVersion)
	pluginPath := d.GetCachedPluginPath(driver, majorVersion)

	// Remove the specific plugin binary
	if err := os.Remove(pluginPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plugin binary %s: %w", pluginPath, err)
	}

	// Try to remove the version directory if it's empty
	if err := os.Remove(pluginDir); err != nil && !os.IsNotExist(err) {
		// Directory might not be empty (other plugins), which is fine
		if !strings.Contains(err.Error(), "directory not empty") {
			return fmt.Errorf("failed to remove plugin directory %s: %w", pluginDir, err)
		}
	}

	return nil
}
