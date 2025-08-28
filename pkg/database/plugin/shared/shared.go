package shared

import (
	"encoding/gob"

	"github.com/hashicorp/go-plugin"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion: 1,

	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "SCHEMAHERO_PLUGIN",
	MagicCookieValue: "schemahero-database-plugin",
}

// PluginMap is the map of plugins we can dispense.
// This will be populated by the plugin package to avoid circular dependencies.
var PluginMap = map[string]plugin.Plugin{}

// Interface name constant for database plugins
const DatabaseInterfaceName = "database"

// RegisterSchemaTypes registers all schema types for gob encoding/decoding
// This should be called once by each plugin to ensure proper serialization
// across the RPC boundary
func RegisterSchemaTypes() {
	// Register core table schema types for all databases
	gob.Register(&schemasv1alpha4.PostgresqlTableSchema{})
	gob.Register(&schemasv1alpha4.MysqlTableSchema{})
	gob.Register(&schemasv1alpha4.TimescaleDBTableSchema{})
	gob.Register(&schemasv1alpha4.CassandraTableSchema{})
	gob.Register(&schemasv1alpha4.SqliteTableSchema{})
	gob.Register(&schemasv1alpha4.RqliteTableSchema{})

	// Register view schema types
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.TimescaleDBViewSchema{})

	// Register function schema types
	gob.Register(&schemasv1alpha4.PostgresqlFunctionSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedFunctionSchema{})

	// Register extension and utility types
	gob.Register(&schemasv1alpha4.PostgresDatabaseExtension{})
	gob.Register(&schemasv1alpha4.SeedData{})

	// Register PostgreSQL nested types
	gob.Register(&schemasv1alpha4.PostgresqlTableColumn{})
	gob.Register(&schemasv1alpha4.PostgresqlTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.PostgresqlTableColumnAttributes{})
	gob.Register(&schemasv1alpha4.PostgresqlTableForeignKey{})
	gob.Register(&schemasv1alpha4.PostgresqlTableIndex{})
	gob.Register(&schemasv1alpha4.PostgresqlTableTrigger{})

	// Register MySQL nested types
	gob.Register(&schemasv1alpha4.MysqlTableColumn{})
	gob.Register(&schemasv1alpha4.MysqlTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.MysqlTableColumnAttributes{})
	gob.Register(&schemasv1alpha4.MysqlTableForeignKey{})
	gob.Register(&schemasv1alpha4.MysqlTableIndex{})

	// Register SQLite nested types
	gob.Register(&schemasv1alpha4.SqliteTableColumn{})
	gob.Register(&schemasv1alpha4.SqliteTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.SqliteTableForeignKey{})
	gob.Register(&schemasv1alpha4.SqliteTableIndex{})

	// Register RQLite nested types
	gob.Register(&schemasv1alpha4.RqliteTableColumn{})
	gob.Register(&schemasv1alpha4.RqliteTableColumnConstraints{})
	gob.Register(&schemasv1alpha4.RqliteTableColumnAttributes{})
	gob.Register(&schemasv1alpha4.RqliteTableForeignKey{})
	gob.Register(&schemasv1alpha4.RqliteTableForeignKeyReferences{})
	gob.Register(&schemasv1alpha4.RqliteTableIndex{})

	// Register Cassandra nested types
	gob.Register(&schemasv1alpha4.CassandraColumn{})
	gob.Register(&schemasv1alpha4.CassandraClusteringOrder{})
	gob.Register(&schemasv1alpha4.CassandraTableProperties{})
	gob.Register(&schemasv1alpha4.CassandraField{})
	gob.Register(&schemasv1alpha4.CassandraDataTypeSchema{})

	// Register TimescaleDB nested types (TimescaleDB reuses PostgreSQL types)
	gob.Register(&schemasv1alpha4.TimescaleDBHypertable{})
	gob.Register(&schemasv1alpha4.TimescaleDBCompression{})
	gob.Register(&schemasv1alpha4.TimescaleDBRetention{})
}
