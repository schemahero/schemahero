package main

import (
	"encoding/gob"
	
	"github.com/hashicorp/go-plugin"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemaheroplugin "github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/plugin/shared"
)

func init() {
	// CRITICAL: Register ALL types that will be passed through RPC
	// This includes schema types and any nested types
	gob.Register(&schemasv1alpha4.CassandraTableSchema{})
	gob.Register(&schemasv1alpha4.CassandraDataTypeSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.SeedData{})
	
	// Register nested types used in CassandraTableSchema
	gob.Register(&schemasv1alpha4.CassandraColumn{})
	gob.Register(&schemasv1alpha4.CassandraClusteringOrder{})
	gob.Register(&schemasv1alpha4.CassandraTableProperties{})
	gob.Register(&schemasv1alpha4.CassandraField{})
}

func main() {
	// Create the cassandra plugin implementation
	cassandraPlugin := &CassandraPlugin{}

	// Create the RPC plugin wrapper
	rpcPlugin := &schemaheroplugin.DatabaseRPCPlugin{
		Impl: cassandraPlugin,
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