package plugin

import (
	"net/rpc"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

// ConnectionProxy implements the SchemaHeroDatabaseConnection interface
// and proxies all method calls over RPC to the actual connection in the plugin process.
// This allows database connections to work across the plugin boundary.
type ConnectionProxy struct {
	// client is the RPC client used to communicate with the plugin server
	client *rpc.Client
	// connectionID uniquely identifies this connection on the server side
	connectionID string
}

// NewConnectionProxy creates a new connection proxy with the given RPC client and connection ID.
func NewConnectionProxy(client *rpc.Client, connectionID string) *ConnectionProxy {
	return &ConnectionProxy{
		client:       client,
		connectionID: connectionID,
	}
}

// Close implements interfaces.SchemaHeroDatabaseConnection.Close()
func (c *ConnectionProxy) Close() error {
	var reply ConnectionCloseReply
	err := c.client.Call("Plugin.ConnectionClose", &ConnectionCloseArgs{
		ConnectionID: c.connectionID,
	}, &reply)
	if err != nil {
		return err
	}

	if reply.Error != "" {
		return &BasicError{Message: reply.Error}
	}

	return nil
}

// DatabaseName implements interfaces.SchemaHeroDatabaseConnection.DatabaseName()
func (c *ConnectionProxy) DatabaseName() string {
	var reply ConnectionDatabaseNameReply
	err := c.client.Call("Plugin.ConnectionDatabaseName", &ConnectionDatabaseNameArgs{
		ConnectionID: c.connectionID,
	}, &reply)
	if err != nil {
		return ""
	}
	return reply.Name
}

// EngineVersion implements interfaces.SchemaHeroDatabaseConnection.EngineVersion()
func (c *ConnectionProxy) EngineVersion() string {
	var reply ConnectionEngineVersionReply
	err := c.client.Call("Plugin.ConnectionEngineVersion", &ConnectionEngineVersionArgs{
		ConnectionID: c.connectionID,
	}, &reply)
	if err != nil {
		return ""
	}
	return reply.Version
}

// ListTables implements interfaces.SchemaHeroDatabaseConnection.ListTables()
func (c *ConnectionProxy) ListTables() ([]*types.Table, error) {
	var reply ConnectionListTablesReply
	err := c.client.Call("Plugin.ConnectionListTables", &ConnectionListTablesArgs{
		ConnectionID: c.connectionID,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Tables, nil
}

// ListTableForeignKeys implements interfaces.SchemaHeroDatabaseConnection.ListTableForeignKeys()
func (c *ConnectionProxy) ListTableForeignKeys(schema, table string) ([]*types.ForeignKey, error) {
	var reply ConnectionListTableForeignKeysReply
	err := c.client.Call("Plugin.ConnectionListTableForeignKeys", &ConnectionListTableForeignKeysArgs{
		ConnectionID: c.connectionID,
		Schema:       schema,
		Table:        table,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.ForeignKeys, nil
}

// ListTableIndexes implements interfaces.SchemaHeroDatabaseConnection.ListTableIndexes()
func (c *ConnectionProxy) ListTableIndexes(schema, table string) ([]*types.Index, error) {
	var reply ConnectionListTableIndexesReply
	err := c.client.Call("Plugin.ConnectionListTableIndexes", &ConnectionListTableIndexesArgs{
		ConnectionID: c.connectionID,
		Schema:       schema,
		Table:        table,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Indexes, nil
}

// GetTablePrimaryKey implements interfaces.SchemaHeroDatabaseConnection.GetTablePrimaryKey()
func (c *ConnectionProxy) GetTablePrimaryKey(table string) (*types.KeyConstraint, error) {
	var reply ConnectionGetTablePrimaryKeyReply
	err := c.client.Call("Plugin.ConnectionGetTablePrimaryKey", &ConnectionGetTablePrimaryKeyArgs{
		ConnectionID: c.connectionID,
		Table:        table,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.PrimaryKey, nil
}

// GetTableSchema implements interfaces.SchemaHeroDatabaseConnection.GetTableSchema()
func (c *ConnectionProxy) GetTableSchema(table string) ([]*types.Column, error) {
	var reply ConnectionGetTableSchemaReply
	err := c.client.Call("Plugin.ConnectionGetTableSchema", &ConnectionGetTableSchemaArgs{
		ConnectionID: c.connectionID,
		Table:        table,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Columns, nil
}

// PlanTableSchema implements interfaces.SchemaHeroDatabaseConnection.PlanTableSchema()
func (c *ConnectionProxy) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// WORKAROUND: gob encoding loses empty string pointers (treats them as nil)
	// Before sending through RPC, replace empty string defaults with a sentinel value
	// that will be restored on the other side
	const emptyStringSentinel = "__SCHEMAHERO_EMPTY_STRING_DEFAULT__"

	// Check if this is a PostgreSQL or MySQL schema and fix empty string defaults
	if pgSchema, ok := tableSchema.(*schemasv1alpha4.PostgresqlTableSchema); ok {
		for _, col := range pgSchema.Columns {
			if col.Default != nil && *col.Default == "" {
				sentinel := emptyStringSentinel
				col.Default = &sentinel
			}
		}
	} else if mysqlSchema, ok := tableSchema.(*schemasv1alpha4.MysqlTableSchema); ok {
		for _, col := range mysqlSchema.Columns {
			if col.Default != nil && *col.Default == "" {
				sentinel := emptyStringSentinel
				col.Default = &sentinel
			}
		}
	}

	var reply ConnectionPlanTableSchemaReply
	err := c.client.Call("Plugin.ConnectionPlanTableSchema", &ConnectionPlanTableSchemaArgs{
		ConnectionID: c.connectionID,
		TableName:    tableName,
		TableSchema:  tableSchema,
		SeedData:     seedData,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Statements, nil
}

// PlanViewSchema implements interfaces.SchemaHeroDatabaseConnection.PlanViewSchema()
func (c *ConnectionProxy) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	var reply ConnectionPlanViewSchemaReply
	err := c.client.Call("Plugin.ConnectionPlanViewSchema", &ConnectionPlanViewSchemaArgs{
		ConnectionID: c.connectionID,
		ViewName:     viewName,
		ViewSchema:   viewSchema,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Statements, nil
}

// PlanFunctionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanFunctionSchema()
func (c *ConnectionProxy) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	var reply ConnectionPlanFunctionSchemaReply
	err := c.client.Call("Plugin.ConnectionPlanFunctionSchema", &ConnectionPlanFunctionSchemaArgs{
		ConnectionID:   c.connectionID,
		FunctionName:   functionName,
		FunctionSchema: functionSchema,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Statements, nil
}

// PlanExtensionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanExtensionSchema()
func (c *ConnectionProxy) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	var reply ConnectionPlanExtensionSchemaReply
	err := c.client.Call("Plugin.ConnectionPlanExtensionSchema", &ConnectionPlanExtensionSchemaArgs{
		ConnectionID:    c.connectionID,
		ExtensionName:   extensionName,
		ExtensionSchema: extensionSchema,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Statements, nil
}

// DeployStatements implements interfaces.SchemaHeroDatabaseConnection.DeployStatements()
func (c *ConnectionProxy) DeployStatements(statements []string) error {
	var reply ConnectionDeployStatementsReply
	err := c.client.Call("Plugin.ConnectionDeployStatements", &ConnectionDeployStatementsArgs{
		ConnectionID: c.connectionID,
		Statements:   statements,
	}, &reply)
	if err != nil {
		return err
	}

	if reply.Error != "" {
		return &BasicError{Message: reply.Error}
	}

	return nil
}

// GenerateFixtures implements interfaces.SchemaHeroDatabaseConnection.GenerateFixtures()
func (c *ConnectionProxy) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	// WORKAROUND: gob encoding loses empty string pointers
	// Apply the same sentinel value replacement for fixtures
	const emptyStringSentinel = "__SCHEMAHERO_EMPTY_STRING_DEFAULT__"

	if spec != nil && spec.Schema != nil {
		if spec.Schema.Postgres != nil {
			for _, col := range spec.Schema.Postgres.Columns {
				if col.Default != nil && *col.Default == "" {
					sentinel := emptyStringSentinel
					col.Default = &sentinel
				}
			}
		} else if spec.Schema.Mysql != nil {
			for _, col := range spec.Schema.Mysql.Columns {
				if col.Default != nil && *col.Default == "" {
					sentinel := emptyStringSentinel
					col.Default = &sentinel
				}
			}
		}
	}

	var reply ConnectionGenerateFixturesReply
	err := c.client.Call("Plugin.ConnectionGenerateFixtures", &ConnectionGenerateFixturesArgs{
		ConnectionID: c.connectionID,
		Spec:         spec,
	}, &reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &BasicError{Message: reply.Error}
	}

	return reply.Statements, nil
}

// Connection RPC method argument and reply types

// ConnectionCloseArgs represents the arguments for the ConnectionClose RPC call.
type ConnectionCloseArgs struct {
	ConnectionID string
}

// ConnectionCloseReply represents the response for the ConnectionClose RPC call.
type ConnectionCloseReply struct {
	Error string
}

// ConnectionDatabaseNameArgs represents the arguments for the ConnectionDatabaseName RPC call.
type ConnectionDatabaseNameArgs struct {
	ConnectionID string
}

// ConnectionDatabaseNameReply represents the response for the ConnectionDatabaseName RPC call.
type ConnectionDatabaseNameReply struct {
	Name string
}

// ConnectionEngineVersionArgs represents the arguments for the ConnectionEngineVersion RPC call.
type ConnectionEngineVersionArgs struct {
	ConnectionID string
}

// ConnectionEngineVersionReply represents the response for the ConnectionEngineVersion RPC call.
type ConnectionEngineVersionReply struct {
	Version string
}

// ConnectionListTablesArgs represents the arguments for the ConnectionListTables RPC call.
type ConnectionListTablesArgs struct {
	ConnectionID string
}

// ConnectionListTablesReply represents the response for the ConnectionListTables RPC call.
type ConnectionListTablesReply struct {
	Tables []*types.Table
	Error  string
}

// ConnectionListTableForeignKeysArgs represents the arguments for the ConnectionListTableForeignKeys RPC call.
type ConnectionListTableForeignKeysArgs struct {
	ConnectionID string
	Schema       string
	Table        string
}

// ConnectionListTableForeignKeysReply represents the response for the ConnectionListTableForeignKeys RPC call.
type ConnectionListTableForeignKeysReply struct {
	ForeignKeys []*types.ForeignKey
	Error       string
}

// ConnectionListTableIndexesArgs represents the arguments for the ConnectionListTableIndexes RPC call.
type ConnectionListTableIndexesArgs struct {
	ConnectionID string
	Schema       string
	Table        string
}

// ConnectionListTableIndexesReply represents the response for the ConnectionListTableIndexes RPC call.
type ConnectionListTableIndexesReply struct {
	Indexes []*types.Index
	Error   string
}

// ConnectionGetTablePrimaryKeyArgs represents the arguments for the ConnectionGetTablePrimaryKey RPC call.
type ConnectionGetTablePrimaryKeyArgs struct {
	ConnectionID string
	Table        string
}

// ConnectionGetTablePrimaryKeyReply represents the response for the ConnectionGetTablePrimaryKey RPC call.
type ConnectionGetTablePrimaryKeyReply struct {
	PrimaryKey *types.KeyConstraint
	Error      string
}

// ConnectionGetTableSchemaArgs represents the arguments for the ConnectionGetTableSchema RPC call.
type ConnectionGetTableSchemaArgs struct {
	ConnectionID string
	Table        string
}

// ConnectionGetTableSchemaReply represents the response for the ConnectionGetTableSchema RPC call.
type ConnectionGetTableSchemaReply struct {
	Columns []*types.Column
	Error   string
}

// ConnectionPlanTableSchemaArgs represents the arguments for the ConnectionPlanTableSchema RPC call.
type ConnectionPlanTableSchemaArgs struct {
	ConnectionID string
	TableName    string
	TableSchema  interface{}
	SeedData     *schemasv1alpha4.SeedData
}

// ConnectionPlanTableSchemaReply represents the response for the ConnectionPlanTableSchema RPC call.
type ConnectionPlanTableSchemaReply struct {
	Statements []string
	Error      string
}

// ConnectionPlanViewSchemaArgs represents the arguments for the ConnectionPlanViewSchema RPC call.
type ConnectionPlanViewSchemaArgs struct {
	ConnectionID string
	ViewName     string
	ViewSchema   interface{}
}

// ConnectionPlanViewSchemaReply represents the response for the ConnectionPlanViewSchema RPC call.
type ConnectionPlanViewSchemaReply struct {
	Statements []string
	Error      string
}

// ConnectionPlanFunctionSchemaArgs represents the arguments for the ConnectionPlanFunctionSchema RPC call.
type ConnectionPlanFunctionSchemaArgs struct {
	ConnectionID   string
	FunctionName   string
	FunctionSchema interface{}
}

// ConnectionPlanFunctionSchemaReply represents the response for the ConnectionPlanFunctionSchema RPC call.
type ConnectionPlanFunctionSchemaReply struct {
	Statements []string
	Error      string
}

// ConnectionPlanExtensionSchemaArgs represents the arguments for the ConnectionPlanExtensionSchema RPC call.
type ConnectionPlanExtensionSchemaArgs struct {
	ConnectionID    string
	ExtensionName   string
	ExtensionSchema interface{}
}

// ConnectionPlanExtensionSchemaReply represents the response for the ConnectionPlanExtensionSchema RPC call.
type ConnectionPlanExtensionSchemaReply struct {
	Statements []string
	Error      string
}

// ConnectionDeployStatementsArgs represents the arguments for the ConnectionDeployStatements RPC call.
type ConnectionDeployStatementsArgs struct {
	ConnectionID string
	Statements   []string
}

// ConnectionDeployStatementsReply represents the response for the ConnectionDeployStatements RPC call.
type ConnectionDeployStatementsReply struct {
	Error string
}

// ConnectionGenerateFixturesArgs represents the arguments for the ConnectionGenerateFixtures RPC call.
type ConnectionGenerateFixturesArgs struct {
	ConnectionID string
	Spec         *schemasv1alpha4.TableSpec
}

// ConnectionGenerateFixturesReply represents the response for the ConnectionGenerateFixtures RPC call.
type ConnectionGenerateFixturesReply struct {
	Statements []string
	Error      string
}

// BasicError represents a basic error that can be transmitted over RPC.
type BasicError struct {
	Message string
}

// Error implements the error interface.
func (e *BasicError) Error() string {
	return e.Message
}
