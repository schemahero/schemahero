package schemaherokubectlcli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schemahero/schemahero/pkg/database/plugin"
)

func TestPluginDownloadCmdUsesCachedPlugin(t *testing.T) {
	plugin.ResetGlobalPluginSystem()
	t.Cleanup(plugin.ResetGlobalPluginSystem)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	pluginDir := filepath.Join(homeDir, ".schemahero", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	pluginPath := filepath.Join(pluginDir, "schemahero-postgres")
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := PluginDownloadCmd()
	cmd.SetArgs([]string{"postgres"})
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected plugin download command to use cached plugin: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Downloaded SchemaHero postgres plugin to ") {
		t.Fatalf("unexpected output: %q", output)
	}

	if !strings.Contains(output, pluginPath) {
		t.Fatalf("expected output to include cached plugin path %q, got %q", pluginPath, output)
	}
}
