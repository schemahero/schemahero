package main

import (
	"github.com/hashicorp/go-plugin"
	schemaheroplugin "github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/plugin/shared"
)

func init() {
	// Register all schema types for RPC serialization
	shared.RegisterSchemaTypes()
}

func main() {
	// Create the timescaledb plugin implementation
	timescaledbPlugin := &TimescaleDBPlugin{}

	// Create the RPC plugin wrapper
	rpcPlugin := &schemaheroplugin.DatabaseRPCPlugin{
		Impl: timescaledbPlugin,
	}

	// Create the plugin map
	pluginMap := map[string]plugin.Plugin{
		"database": rpcPlugin,
	}

	// Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
	})
}