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
	gob.Register(&schemasv1alpha4.PostgresqlTableSchema{})
	gob.Register(&schemasv1alpha4.MysqlTableSchema{})
	gob.Register(&schemasv1alpha4.TimescaleDBTableSchema{})
	gob.Register(&schemasv1alpha4.CassandraTableSchema{})
	gob.Register(&schemasv1alpha4.SqliteTableSchema{})
	gob.Register(&schemasv1alpha4.RqliteTableSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.PostgresqlFunctionSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedFunctionSchema{})
	gob.Register(&schemasv1alpha4.PostgresDatabaseExtension{})
	gob.Register(&schemasv1alpha4.SeedData{})
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