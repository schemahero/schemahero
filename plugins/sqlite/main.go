package main

import (
	"encoding/gob"
	
	"github.com/hashicorp/go-plugin"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemaheroplugin "github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/plugin/shared"
)

func init() {
	// Register types that will be passed through RPC interfaces
	// These are needed for gob encoding/decoding of interface{} parameters
	gob.Register(&schemasv1alpha4.SqliteTableSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.SeedData{})
	
	// Register nested types used in SqliteTableSchema
	gob.Register(&schemasv1alpha4.SqliteTableColumn{})
	gob.Register(&schemasv1alpha4.SqliteTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.SqliteTableForeignKey{})
	gob.Register(&schemasv1alpha4.SqliteTableIndex{})
}

func main() {
	// Create the sqlite plugin implementation
	sqlitePlugin := &SqlitePlugin{}

	// Create the RPC plugin wrapper
	rpcPlugin := &schemaheroplugin.DatabaseRPCPlugin{
		Impl: sqlitePlugin,
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