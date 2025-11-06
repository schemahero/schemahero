package plugin

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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
	return fmt.Sprintf("docker.io/schemahero/plugin-%s:0", driver)
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

	// Create repository reference (without tag)
	repoURL := strings.TrimSuffix(artifactRef, ":"+tag)
	repo, err := remote.NewRepository(repoURL)
	if err != nil {
		return fmt.Errorf("failed to create repository reference for %s: %w", repoURL, err)
	}

	// Use file store with higher size limit (100MB) for plugin artifacts
	fs, err := file.NewWithFallbackLimit(cacheDir, 100*1024*1024) // 100MB limit
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	// Use default copy options for custom OCI artifacts
	copyOpts := oras.CopyOptions{}

	_, err = oras.Copy(ctx, repo, tag, fs, tag, copyOpts)
	if err != nil {
		return fmt.Errorf("failed to download plugin from %s (repo: %s, tag: %s): %w", artifactRef, repoURL, tag, err)
	}
	
	// ORAS downloads all platform artifacts, so we must check for the correct platform-specific tarball FIRST
	// before falling back to generic paths
	possiblePaths := []string{
		// Look for the OS-specific tarball first (runtime.GOOS and runtime.GOARCH determine the correct platform)
		filepath.Join(cacheDir, fmt.Sprintf("schemahero-%s-%s-%s.tar.gz", driver, runtime.GOOS, runtime.GOARCH)),
		// Fallback paths for direct binary (no tarball)
		filepath.Join(cacheDir, fmt.Sprintf("schemahero-%s", driver)),           // Direct
		filepath.Join(cacheDir, "plugins", "bin", fmt.Sprintf("schemahero-%s", driver)), // With structure
		filepath.Join(cacheDir, "plugins", fmt.Sprintf("schemahero-%s", driver)),        // Partial structure
	}
	
	var foundPath string
	var isArchive bool
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			isArchive = strings.HasSuffix(path, ".tar.gz")
			break
		}
	}
	
	if foundPath == "" {
		return fmt.Errorf("plugin not found in any expected location after download")
	}
	
	var finalPluginPath string
	
	if isArchive {
		// Extract the tarball
		finalPluginPath, err = d.extractPlugin(foundPath, cacheDir, driver)
		if err != nil {
			return fmt.Errorf("failed to extract plugin from %s: %w", foundPath, err)
		}
	} else {
		finalPluginPath = foundPath
	}

	// Move to final location if needed
	if finalPluginPath != pluginPath {
		if err := os.Rename(finalPluginPath, pluginPath); err != nil {
			return fmt.Errorf("failed to move plugin from %s to %s: %w", finalPluginPath, pluginPath, err)
		}
	}

	// Ensure the binary is executable
	if err := os.Chmod(pluginPath, 0755); err != nil {
		return fmt.Errorf("failed to make plugin binary executable: %w", err)
	}

	return nil
}

// extractPlugin extracts a plugin binary from a tar.gz archive
func (d *PluginDownloader) extractPlugin(archivePath, extractDir, driver string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	
	expectedBinary := fmt.Sprintf("schemahero-%s", driver)
	
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Sanitize the archive entry name to prevent Zip Slip
		if strings.Contains(header.Name, "..") || strings.Contains(header.Name, "/") || strings.Contains(header.Name, "\\") {
			// Unsafe entry, skip extraction
			continue
		}

		// Look for the binary we want
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, expectedBinary) {
			extractPath := filepath.Join(extractDir, filepath.Base(header.Name))
			
			outFile, err := os.Create(extractPath)
			if err != nil {
				return "", fmt.Errorf("failed to create extract file: %w", err)
			}
			
			_, err = io.Copy(outFile, tr)
			outFile.Close()
			
			if err != nil {
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			
			// Make it executable
			if err := os.Chmod(extractPath, 0755); err != nil {
				return "", fmt.Errorf("failed to make extracted binary executable: %w", err)
			}
			
			return extractPath, nil
		}
	}
	
	return "", fmt.Errorf("binary %s not found in archive", expectedBinary)
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
