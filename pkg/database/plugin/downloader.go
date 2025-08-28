package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

// NewPluginDownloader creates a new plugin downloader with specified cache directory
func NewPluginDownloader(cacheDir string) *PluginDownloader {
	if cacheDir == "" {
		cacheDir = "/tmp/plugins"
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
func (d *PluginDownloader) GetPluginArtifactRef(driver string, majorVersion string) string {
	return fmt.Sprintf("schemahero/plugin-%s:%s", driver, majorVersion)
}

// GetCachedPluginPath returns the expected path for a cached plugin binary
func (d *PluginDownloader) GetCachedPluginPath(driver string, majorVersion string) string {
	pluginBinary := fmt.Sprintf("schemahero-%s", driver)
	return filepath.Join(d.cacheDir, majorVersion, pluginBinary)
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

	// Create file store for download
	fs, err := file.New(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	// Get the current platform
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	
	// Copy from registry to local file store
	// ORAS will automatically select the correct platform from multi-arch image
	desc, err := oras.Copy(ctx, repo, "latest", fs, "latest", oras.CopyOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			// This ensures we get the right platform from multi-arch
			PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				// Only copy if it matches our platform or is the manifest
				if desc.Platform != nil {
					descPlatform := fmt.Sprintf("%s/%s", desc.Platform.OS, desc.Platform.Architecture)
					if descPlatform != platform {
						return oras.SkipNode
					}
				}
				return nil
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download plugin %s: %w", artifactRef, err)
	}

	// Find the downloaded plugin binary
	// ORAS downloads files maintaining their structure, so we need to find the binary
	downloadedFiles := []string{}
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Look for files that start with "schemahero-" and are executable
		if !info.IsDir() && strings.HasPrefix(info.Name(), "schemahero-") {
			downloadedFiles = append(downloadedFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to search for downloaded plugin binary: %w", err)
	}

	if len(downloadedFiles) == 0 {
		return fmt.Errorf("no plugin binary found after downloading %s (descriptor: %v)", artifactRef, desc)
	}

	// Use the first matching file (there should only be one)
	downloadedPath := downloadedFiles[0]
	
	// Move to expected location if it's not already there
	if downloadedPath != pluginPath {
		if err := os.Rename(downloadedPath, pluginPath); err != nil {
			return fmt.Errorf("failed to move plugin binary to expected location: %w", err)
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