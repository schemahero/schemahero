package main

import (
	"github.com/hashicorp/go-plugin"
	schemaheroplugin "github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/plugin/shared"
)

// Multi-platform support: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64

func init() {
	// Register all schema types for RPC serialization
	shared.RegisterSchemaTypes()
}

func main() {
	// Create the mysql plugin implementation
	mysqlPlugin := &MySQLPlugin{}

	// Create the RPC plugin wrapper
	rpcPlugin := &schemaheroplugin.DatabaseRPCPlugin{
		Impl: mysqlPlugin,
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