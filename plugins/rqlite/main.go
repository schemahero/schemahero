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
	gob.Register(&schemasv1alpha4.RqliteTableSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.SeedData{})
	
	// Register nested types used in RqliteTableSchema
	gob.Register(&schemasv1alpha4.RqliteTableColumn{})
	gob.Register(&schemasv1alpha4.RqliteTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.RqliteTableColumnAttributes{})
	gob.Register(&schemasv1alpha4.RqliteTableForeignKey{})
	gob.Register(&schemasv1alpha4.RqliteTableForeignKeyReferences{})
	gob.Register(&schemasv1alpha4.RqliteTableIndex{})
}

func main() {
	// Create the rqlite plugin implementation
	rqlitePlugin := &RqlitePlugin{}

	// Create the RPC plugin wrapper
	rpcPlugin := &schemaheroplugin.DatabaseRPCPlugin{
		Impl: rqlitePlugin,
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