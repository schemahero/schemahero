package plugin

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/interfaces"
)

func init() {
	// Register types that will be passed through RPC interfaces
	// These are needed for gob encoding/decoding of interface{} parameters
	gob.Register(&schemasv1alpha4.PostgresqlTableSchema{})
	gob.Register(&schemasv1alpha4.PostgresqlTableColumn{})
	gob.Register(&schemasv1alpha4.MysqlTableSchema{})
	gob.Register(&schemasv1alpha4.TimescaleDBTableSchema{})
	gob.Register(&schemasv1alpha4.CassandraTableSchema{})
	gob.Register(&schemasv1alpha4.CassandraDataTypeSchema{})
	// Register nested types used in CassandraTableSchema
	gob.Register(&schemasv1alpha4.CassandraColumn{})
	gob.Register(&schemasv1alpha4.CassandraClusteringOrder{})
	gob.Register(&schemasv1alpha4.CassandraTableProperties{})
	gob.Register(&schemasv1alpha4.CassandraField{})
	gob.Register(&schemasv1alpha4.SqliteTableSchema{})
	gob.Register(&schemasv1alpha4.RqliteTableSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedViewSchema{})
	gob.Register(&schemasv1alpha4.TimescaleDBViewSchema{})
	gob.Register(&schemasv1alpha4.PostgresqlFunctionSchema{})
	gob.Register(&schemasv1alpha4.NotImplementedFunctionSchema{})
	gob.Register(&schemasv1alpha4.PostgresDatabaseExtension{})
	gob.Register(&schemasv1alpha4.SeedData{})
}

// DatabaseRPCPlugin implements the plugin.Plugin interface for database plugins.
// It handles the creation of RPC server and client instances for the DatabasePlugin interface.
type DatabaseRPCPlugin struct {
	// Impl is the actual implementation of the DatabasePlugin interface.
	// This field is only used on the server side.
	Impl DatabasePlugin
}

// Server returns a new RPC server that serves the DatabasePlugin interface.
// This method is called on the plugin side to create a server that will handle
// requests from the host process.
func (p *DatabaseRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{
		Impl:        p.Impl,
		connections: make(map[string]interfaces.SchemaHeroDatabaseConnection),
	}, nil
}

// Client returns a client implementation that communicates with the plugin server.
// This method is called on the host side to create a client that can make RPC calls
// to the plugin server.
func (p *DatabaseRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// RPCServer wraps a DatabasePlugin implementation and provides RPC methods
// that can be called remotely. This struct acts as the server-side adapter
// between the RPC calls and the actual plugin implementation.
type RPCServer struct {
	// Impl is the concrete DatabasePlugin implementation that this server wraps.
	Impl DatabasePlugin
	// connections stores active database connections indexed by connection ID
	connections map[string]interfaces.SchemaHeroDatabaseConnection
	// connectionsMutex protects concurrent access to the connections map
	connectionsMutex sync.RWMutex
	// nextConnectionID is used to generate unique connection IDs
	nextConnectionID int
}

// NameArgs represents the arguments for the Name RPC call.
type NameArgs struct{}

// NameReply represents the response for the Name RPC call.
type NameReply struct {
	Name string
}

// Name handles RPC calls for the DatabasePlugin.Name() method.
func (s *RPCServer) Name(args *NameArgs, reply *NameReply) error {
	reply.Name = s.Impl.Name()
	return nil
}

// VersionArgs represents the arguments for the Version RPC call.
type VersionArgs struct{}

// VersionReply represents the response for the Version RPC call.
type VersionReply struct {
	Version string
}

// Version handles RPC calls for the DatabasePlugin.Version() method.
func (s *RPCServer) Version(args *VersionArgs, reply *VersionReply) error {
	reply.Version = s.Impl.Version()
	return nil
}

// SupportedEnginesArgs represents the arguments for the SupportedEngines RPC call.
type SupportedEnginesArgs struct{}

// SupportedEnginesReply represents the response for the SupportedEngines RPC call.
type SupportedEnginesReply struct {
	Engines []string
}

// SupportedEngines handles RPC calls for the DatabasePlugin.SupportedEngines() method.
func (s *RPCServer) SupportedEngines(args *SupportedEnginesArgs, reply *SupportedEnginesReply) error {
	reply.Engines = s.Impl.SupportedEngines()
	return nil
}

// ConnectArgs represents the arguments for the Connect RPC call.
type ConnectArgs struct {
	URI     string
	Options map[string]interface{}
}

// ConnectReply represents the response for the Connect RPC call.
type ConnectReply struct {
	Success      bool
	Error        string
	ConnectionID string
}

// Connect handles RPC calls for the DatabasePlugin.Connect() method.
// Creates a connection and stores it with a unique ID for future RPC calls.
func (s *RPCServer) Connect(args *ConnectArgs, reply *ConnectReply) error {
	conn, err := s.Impl.Connect(args.URI, args.Options)
	if err != nil {
		reply.Success = false
		reply.Error = err.Error()
		return nil
	}

	s.connectionsMutex.Lock()
	s.nextConnectionID++
	connectionID := fmt.Sprintf("conn-%d", s.nextConnectionID)
	s.connections[connectionID] = conn
	s.connectionsMutex.Unlock()

	reply.Success = true
	reply.ConnectionID = connectionID
	return nil
}

// ValidateArgs represents the arguments for the Validate RPC call.
type ValidateArgs struct {
	Config map[string]interface{}
}

// ValidateReply represents the response for the Validate RPC call.
type ValidateReply struct {
	Valid bool
	Error string
}

// Validate handles RPC calls for the DatabasePlugin.Validate() method.
func (s *RPCServer) Validate(args *ValidateArgs, reply *ValidateReply) error {
	err := s.Impl.Validate(args.Config)
	if err != nil {
		reply.Valid = false
		reply.Error = err.Error()
		return nil
	}
	reply.Valid = true
	return nil
}

// InitializeArgs represents the arguments for the Initialize RPC call.
type InitializeArgs struct {
	// Context cannot be serialized, so we'll handle it on the server side
}

// InitializeReply represents the response for the Initialize RPC call.
type InitializeReply struct {
	Success bool
	Error   string
}

// Initialize handles RPC calls for the DatabasePlugin.Initialize() method.
func (s *RPCServer) Initialize(args *InitializeArgs, reply *InitializeReply) error {
	// Create a background context since we can't serialize the original context
	ctx := context.Background()
	err := s.Impl.Initialize(ctx)
	if err != nil {
		reply.Success = false
		reply.Error = err.Error()
		return nil
	}
	reply.Success = true
	return nil
}

// ShutdownArgs represents the arguments for the Shutdown RPC call.
type ShutdownArgs struct {
	// Context cannot be serialized, so we'll handle it on the server side
}

// ShutdownReply represents the response for the Shutdown RPC call.
type ShutdownReply struct {
	Success bool
	Error   string
}

// Shutdown handles RPC calls for the DatabasePlugin.Shutdown() method.
func (s *RPCServer) Shutdown(args *ShutdownArgs, reply *ShutdownReply) error {
	// Create a background context since we can't serialize the original context
	ctx := context.Background()
	err := s.Impl.Shutdown(ctx)
	if err != nil {
		reply.Success = false
		reply.Error = err.Error()
		return nil
	}
	reply.Success = true
	return nil
}

// Connection RPC methods for the server

// ConnectionClose handles RPC calls for closing a database connection.
func (s *RPCServer) ConnectionClose(args *ConnectionCloseArgs, reply *ConnectionCloseReply) error {
	s.connectionsMutex.Lock()
	conn, exists := s.connections[args.ConnectionID]
	if exists {
		delete(s.connections, args.ConnectionID)
	}
	s.connectionsMutex.Unlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	if err := conn.Close(); err != nil {
		reply.Error = err.Error()
	}

	return nil
}

// ConnectionDatabaseName handles RPC calls for getting the database name.
func (s *RPCServer) ConnectionDatabaseName(args *ConnectionDatabaseNameArgs, reply *ConnectionDatabaseNameReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Name = ""
		return nil
	}

	reply.Name = conn.DatabaseName()
	return nil
}

// ConnectionEngineVersion handles RPC calls for getting the engine version.
func (s *RPCServer) ConnectionEngineVersion(args *ConnectionEngineVersionArgs, reply *ConnectionEngineVersionReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Version = ""
		return nil
	}

	reply.Version = conn.EngineVersion()
	return nil
}

// ConnectionListTables handles RPC calls for listing tables.
func (s *RPCServer) ConnectionListTables(args *ConnectionListTablesArgs, reply *ConnectionListTablesReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	tables, err := conn.ListTables()
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Tables = tables
	return nil
}

// ConnectionListTableForeignKeys handles RPC calls for listing table foreign keys.
func (s *RPCServer) ConnectionListTableForeignKeys(args *ConnectionListTableForeignKeysArgs, reply *ConnectionListTableForeignKeysReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	foreignKeys, err := conn.ListTableForeignKeys(args.Schema, args.Table)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.ForeignKeys = foreignKeys
	return nil
}

// ConnectionListTableIndexes handles RPC calls for listing table indexes.
func (s *RPCServer) ConnectionListTableIndexes(args *ConnectionListTableIndexesArgs, reply *ConnectionListTableIndexesReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	indexes, err := conn.ListTableIndexes(args.Schema, args.Table)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Indexes = indexes
	return nil
}

// ConnectionGetTablePrimaryKey handles RPC calls for getting a table's primary key.
func (s *RPCServer) ConnectionGetTablePrimaryKey(args *ConnectionGetTablePrimaryKeyArgs, reply *ConnectionGetTablePrimaryKeyReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	primaryKey, err := conn.GetTablePrimaryKey(args.Table)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.PrimaryKey = primaryKey
	return nil
}

// ConnectionGetTableSchema handles RPC calls for getting a table's schema.
func (s *RPCServer) ConnectionGetTableSchema(args *ConnectionGetTableSchemaArgs, reply *ConnectionGetTableSchemaReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	columns, err := conn.GetTableSchema(args.Table)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Columns = columns
	return nil
}

// ConnectionPlanTableSchema handles RPC calls for planning table schema changes.
func (s *RPCServer) ConnectionPlanTableSchema(args *ConnectionPlanTableSchemaArgs, reply *ConnectionPlanTableSchemaReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	// WORKAROUND: Restore empty string defaults that were replaced with sentinel values
	// to work around gob encoding losing empty string pointers
	const emptyStringSentinel = "__SCHEMAHERO_EMPTY_STRING_DEFAULT__"

	// Check if this is a PostgreSQL, MySQL, or SQLite schema and restore empty string defaults
	if pgSchema, ok := args.TableSchema.(*schemasv1alpha4.PostgresqlTableSchema); ok {
		for _, col := range pgSchema.Columns {
			if col.Default != nil && *col.Default == emptyStringSentinel {
				emptyStr := ""
				col.Default = &emptyStr
			}
		}
	} else if mysqlSchema, ok := args.TableSchema.(*schemasv1alpha4.MysqlTableSchema); ok {
		for _, col := range mysqlSchema.Columns {
			if col.Default != nil && *col.Default == emptyStringSentinel {
				emptyStr := ""
				col.Default = &emptyStr
			}
		}
	} else if sqliteSchema, ok := args.TableSchema.(*schemasv1alpha4.SqliteTableSchema); ok {
		for _, col := range sqliteSchema.Columns {
			if col.Default != nil && *col.Default == emptyStringSentinel {
				emptyStr := ""
				col.Default = &emptyStr
			}
		}
	}

	statements, err := conn.PlanTableSchema(args.TableName, args.TableSchema, args.SeedData)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Statements = statements
	return nil
}

// ConnectionPlanViewSchema handles RPC calls for planning view schema changes.
func (s *RPCServer) ConnectionPlanViewSchema(args *ConnectionPlanViewSchemaArgs, reply *ConnectionPlanViewSchemaReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	statements, err := conn.PlanViewSchema(args.ViewName, args.ViewSchema)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Statements = statements
	return nil
}

// ConnectionPlanFunctionSchema handles RPC calls for planning function schema changes.
func (s *RPCServer) ConnectionPlanFunctionSchema(args *ConnectionPlanFunctionSchemaArgs, reply *ConnectionPlanFunctionSchemaReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	statements, err := conn.PlanFunctionSchema(args.FunctionName, args.FunctionSchema)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Statements = statements
	return nil
}

// ConnectionPlanExtensionSchema handles RPC calls for planning extension schema changes.
func (s *RPCServer) ConnectionPlanExtensionSchema(args *ConnectionPlanExtensionSchemaArgs, reply *ConnectionPlanExtensionSchemaReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	statements, err := conn.PlanExtensionSchema(args.ExtensionName, args.ExtensionSchema)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Statements = statements
	return nil
}

// ConnectionDeployStatements handles RPC calls for deploying SQL statements.
func (s *RPCServer) ConnectionDeployStatements(args *ConnectionDeployStatementsArgs, reply *ConnectionDeployStatementsReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	err := conn.DeployStatements(args.Statements)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	return nil
}

// ConnectionGenerateFixtures handles RPC calls for generating fixture statements.
func (s *RPCServer) ConnectionGenerateFixtures(args *ConnectionGenerateFixturesArgs, reply *ConnectionGenerateFixturesReply) error {
	s.connectionsMutex.RLock()
	conn, exists := s.connections[args.ConnectionID]
	s.connectionsMutex.RUnlock()

	if !exists {
		reply.Error = fmt.Sprintf("connection %s not found", args.ConnectionID)
		return nil
	}

	// WORKAROUND: Restore empty string defaults that were replaced with sentinel values
	// to work around gob encoding losing empty string pointers
	const emptyStringSentinel = "__SCHEMAHERO_EMPTY_STRING_DEFAULT__"

	// Check if this spec has a PostgreSQL, MySQL, or SQLite schema and restore empty string defaults
	if args.Spec != nil && args.Spec.Schema != nil {
		if args.Spec.Schema.Postgres != nil {
			for _, col := range args.Spec.Schema.Postgres.Columns {
				if col.Default != nil && *col.Default == emptyStringSentinel {
					emptyStr := ""
					col.Default = &emptyStr
				}
			}
		} else if args.Spec.Schema.Mysql != nil {
			for _, col := range args.Spec.Schema.Mysql.Columns {
				if col.Default != nil && *col.Default == emptyStringSentinel {
					emptyStr := ""
					col.Default = &emptyStr
				}
			}
		} else if args.Spec.Schema.SQLite != nil {
			for _, col := range args.Spec.Schema.SQLite.Columns {
				if col.Default != nil && *col.Default == emptyStringSentinel {
					emptyStr := ""
					col.Default = &emptyStr
				}
			}
		}
	}

	statements, err := conn.GenerateFixtures(args.Spec)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Statements = statements
	return nil
}

// RPCClient wraps an RPC client and implements the DatabasePlugin interface.
// This struct acts as the client-side adapter that makes RPC calls to the
// plugin server and translates them back to the DatabasePlugin interface.
type RPCClient struct {
	// client is the underlying RPC client used to communicate with the plugin server.
	client *rpc.Client
}

// Name implements DatabasePlugin.Name() by making an RPC call to the server.
func (c *RPCClient) Name() string {
	var reply NameReply
	err := c.client.Call("Plugin.Name", &NameArgs{}, &reply)
	if err != nil {
		return ""
	}
	return reply.Name
}

// Version implements DatabasePlugin.Version() by making an RPC call to the server.
func (c *RPCClient) Version() string {
	var reply VersionReply
	err := c.client.Call("Plugin.Version", &VersionArgs{}, &reply)
	if err != nil {
		return ""
	}
	return reply.Version
}

// SupportedEngines implements DatabasePlugin.SupportedEngines() by making an RPC call to the server.
func (c *RPCClient) SupportedEngines() []string {
	var reply SupportedEnginesReply
	err := c.client.Call("Plugin.SupportedEngines", &SupportedEnginesArgs{}, &reply)
	if err != nil {
		return nil
	}
	return reply.Engines
}

// Connect implements DatabasePlugin.Connect() by making an RPC call to the server.
// Returns a ConnectionProxy that handles database operations over RPC.
func (c *RPCClient) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	var reply ConnectReply
	err := c.client.Call("Plugin.Connect", &ConnectArgs{
		URI:     uri,
		Options: options,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if !reply.Success {
		return nil, &plugin.BasicError{Message: reply.Error}
	}

	// Return a connection proxy that handles database operations over RPC
	return NewConnectionProxy(c.client, reply.ConnectionID), nil
}

// Validate implements DatabasePlugin.Validate() by making an RPC call to the server.
func (c *RPCClient) Validate(config map[string]interface{}) error {
	var reply ValidateReply
	err := c.client.Call("Plugin.Validate", &ValidateArgs{Config: config}, &reply)
	if err != nil {
		return err
	}

	if !reply.Valid {
		return &plugin.BasicError{Message: reply.Error}
	}

	return nil
}

// Initialize implements DatabasePlugin.Initialize() by making an RPC call to the server.
func (c *RPCClient) Initialize(ctx context.Context) error {
	var reply InitializeReply
	err := c.client.Call("Plugin.Initialize", &InitializeArgs{}, &reply)
	if err != nil {
		return err
	}

	if !reply.Success {
		return &plugin.BasicError{Message: reply.Error}
	}

	return nil
}

// Shutdown implements DatabasePlugin.Shutdown() by making an RPC call to the server.
func (c *RPCClient) Shutdown(ctx context.Context) error {
	var reply ShutdownReply
	err := c.client.Call("Plugin.Shutdown", &ShutdownArgs{}, &reply)
	if err != nil {
		return err
	}

	if !reply.Success {
		return &plugin.BasicError{Message: reply.Error}
	}

	return nil
}
