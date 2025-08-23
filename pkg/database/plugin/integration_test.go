package plugin

import (
	"context"
	"errors"
	"fmt"
	"net/rpc"
	"testing"

	"github.com/hashicorp/go-plugin"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/interfaces"
	"github.com/schemahero/schemahero/pkg/database/types"
)

// TestPlugin implements DatabasePlugin for testing
type TestPlugin struct {
	name             string
	version          string
	supportedEngines []string
	initialized      bool
	connectError     error
	validateError    error
	initError        error
	shutdownError    error
}

func NewTestPlugin() *TestPlugin {
	return &TestPlugin{
		name:             "test-plugin",
		version:          "1.0.0",
		supportedEngines: []string{"testdb", "mockdb"},
		initialized:      false,
	}
}

func (p *TestPlugin) Name() string {
	return p.name
}

func (p *TestPlugin) Version() string {
	return p.version
}

func (p *TestPlugin) SupportedEngines() []string {
	return p.supportedEngines
}

func (p *TestPlugin) Connect(uri string, options map[string]interface{}) (interfaces.SchemaHeroDatabaseConnection, error) {
	if p.connectError != nil {
		return nil, p.connectError
	}
	return NewTestConnection(uri, options), nil
}

func (p *TestPlugin) Validate(config map[string]interface{}) error {
	return p.validateError
}

func (p *TestPlugin) Initialize(ctx context.Context) error {
	if p.initError != nil {
		return p.initError
	}
	p.initialized = true
	return nil
}

func (p *TestPlugin) Shutdown(ctx context.Context) error {
	if p.shutdownError != nil {
		return p.shutdownError
	}
	p.initialized = false
	return nil
}

// Test helper methods
func (p *TestPlugin) SetConnectError(err error) {
	p.connectError = err
}

func (p *TestPlugin) SetValidateError(err error) {
	p.validateError = err
}

func (p *TestPlugin) SetInitError(err error) {
	p.initError = err
}

func (p *TestPlugin) SetShutdownError(err error) {
	p.shutdownError = err
}

func (p *TestPlugin) IsInitialized() bool {
	return p.initialized
}

// TestConnection implements SchemaHeroDatabaseConnection for testing
type TestConnection struct {
	uri         string
	options     map[string]interface{}
	dbName      string
	version     string
	closed      bool
	tables      []*types.Table
	foreignKeys []*types.ForeignKey
	indexes     []*types.Index
	primaryKey  *types.KeyConstraint
	columns     []*types.Column
}

func NewTestConnection(uri string, options map[string]interface{}) *TestConnection {
	return &TestConnection{
		uri:     uri,
		options: options,
		dbName:  "test_database",
		version: "1.2.3",
		tables: []*types.Table{
			{Name: "users", Schema: "public", Charset: "utf8", Collation: "utf8_general_ci"},
			{Name: "orders", Schema: "public", Charset: "utf8", Collation: "utf8_general_ci"},
		},
		foreignKeys: []*types.ForeignKey{
			{
				ChildColumns:  []string{"user_id"},
				ParentTable:   "users",
				ParentColumns: []string{"id"},
				Name:          "fk_orders_user_id",
				OnDelete:      "CASCADE",
			},
		},
		indexes: []*types.Index{
			{Name: "idx_users_email", Columns: []string{"email"}, IsUnique: true},
			{Name: "idx_orders_date", Columns: []string{"created_at"}, IsUnique: false},
		},
		primaryKey: &types.KeyConstraint{
			Name:      "users_pkey",
			Columns:   []string{"id"},
			IsPrimary: true,
		},
		columns: []*types.Column{
			{
				Name:     "id",
				DataType: "integer",
				Constraints: &types.ColumnConstraints{
					NotNull: func() *bool { b := true; return &b }(),
				},
				Attributes: &types.ColumnAttributes{
					AutoIncrement: func() *bool { b := true; return &b }(),
				},
			},
			{
				Name:        "email",
				DataType:    "varchar(255)",
				Constraints: &types.ColumnConstraints{NotNull: func() *bool { b := true; return &b }()},
			},
		},
	}
}

func (c *TestConnection) Close() error {
	if c.closed {
		return errors.New("connection already closed")
	}
	c.closed = true
	return nil
}

func (c *TestConnection) DatabaseName() string {
	return c.dbName
}

func (c *TestConnection) EngineVersion() string {
	return c.version
}

func (c *TestConnection) ListTables() ([]*types.Table, error) {
	if c.closed {
		return nil, errors.New("connection is closed")
	}
	return c.tables, nil
}

func (c *TestConnection) ListTableForeignKeys(schema, table string) ([]*types.ForeignKey, error) {
	if c.closed {
		return nil, errors.New("connection is closed")
	}
	return c.foreignKeys, nil
}

func (c *TestConnection) ListTableIndexes(schema, table string) ([]*types.Index, error) {
	if c.closed {
		return nil, errors.New("connection is closed")
	}
	return c.indexes, nil
}

func (c *TestConnection) GetTablePrimaryKey(table string) (*types.KeyConstraint, error) {
	if c.closed {
		return nil, errors.New("connection is closed")
	}
	return c.primaryKey, nil
}

func (c *TestConnection) GetTableSchema(table string) ([]*types.Column, error) {
	if c.closed {
		return nil, errors.New("connection is closed")
	}
	return c.columns, nil
}

// Test helper methods
func (c *TestConnection) IsClosed() bool {
	return c.closed
}

// PlanTableSchema implements interfaces.SchemaHeroDatabaseConnection.PlanTableSchema()
func (c *TestConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Test implementation - return sample statements
	return []string{fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY)", tableName)}, nil
}

// PlanViewSchema implements interfaces.SchemaHeroDatabaseConnection.PlanViewSchema()
func (c *TestConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	// Test implementation - return sample statements
	return []string{fmt.Sprintf("CREATE VIEW %s AS SELECT * FROM test", viewName)}, nil
}

// PlanFunctionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanFunctionSchema()
func (c *TestConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	// Test implementation - return sample statements
	return []string{fmt.Sprintf("CREATE FUNCTION %s() RETURNS INT", functionName)}, nil
}

// PlanExtensionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanExtensionSchema()
func (c *TestConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	// Test implementation - return sample statements
	return []string{fmt.Sprintf("CREATE EXTENSION %s", extensionName)}, nil
}

// DeployStatements implements interfaces.SchemaHeroDatabaseConnection.DeployStatements()
func (c *TestConnection) DeployStatements(statements []string) error {
	// Test implementation - just return success
	return nil
}

// createInProcessRPCClientServer creates a client/server pair for testing RPC communication in-process
func createInProcessRPCClientServer(impl DatabasePlugin) (*RPCClient, *RPCServer, error) {
	// Create server
	server := &RPCServer{
		Impl:        impl,
		connections: make(map[string]interfaces.SchemaHeroDatabaseConnection),
	}

	// Create a pair of connected pipes to simulate network communication
	serverConn, clientConn := plugin.TestConn(&testing.T{})

	// Start RPC server
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("Plugin", server)
	go rpcServer.ServeConn(serverConn)

	// Create RPC client
	rpcClient := rpc.NewClient(clientConn)
	client := &RPCClient{client: rpcClient}

	return client, server, nil
}

func TestRPCCommunication(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, server, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	t.Run("Name", func(t *testing.T) {
		name := client.Name()
		if name != "test-plugin" {
			t.Errorf("Expected name 'test-plugin', got '%s'", name)
		}
	})

	t.Run("Version", func(t *testing.T) {
		version := client.Version()
		if version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", version)
		}
	})

	t.Run("SupportedEngines", func(t *testing.T) {
		engines := client.SupportedEngines()
		expected := []string{"testdb", "mockdb"}
		if len(engines) != len(expected) {
			t.Errorf("Expected %d engines, got %d", len(expected), len(engines))
		}
		for i, engine := range engines {
			if engine != expected[i] {
				t.Errorf("Expected engine '%s', got '%s'", expected[i], engine)
			}
		}
	})

	// Test error handling - when RPC client is closed, should return empty values
	client.client.Close()
	if client.Name() != "" {
		t.Error("Expected empty name when RPC client is closed")
	}
	if client.Version() != "" {
		t.Error("Expected empty version when RPC client is closed")
	}
	if client.SupportedEngines() != nil {
		t.Error("Expected nil engines when RPC client is closed")
	}

	_ = server // Keep server reference to avoid unused variable
}

func TestRPCConnection(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	t.Run("SuccessfulConnect", func(t *testing.T) {
		uri := "testdb://localhost:5432/testdb"
		options := map[string]interface{}{
			"timeout": 30,
			"ssl":     true,
		}

		conn, err := client.Connect(uri, options)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Verify connection is a proxy
		proxy, ok := conn.(*ConnectionProxy)
		if !ok {
			t.Fatal("Expected ConnectionProxy, got different type")
		}

		// Verify proxy has connection ID
		if proxy.connectionID == "" {
			t.Error("Expected non-empty connection ID")
		}

		// Test connection methods through proxy
		dbName := proxy.DatabaseName()
		if dbName != "test_database" {
			t.Errorf("Expected database name 'test_database', got '%s'", dbName)
		}

		version := proxy.EngineVersion()
		if version != "1.2.3" {
			t.Errorf("Expected engine version '1.2.3', got '%s'", version)
		}
	})

	t.Run("ConnectError", func(t *testing.T) {
		testPlugin.SetConnectError(errors.New("connection failed"))

		_, err := client.Connect("invalid://uri", nil)
		if err == nil {
			t.Error("Expected connection error")
		}

		if err.Error() != "connection failed" {
			t.Errorf("Expected error message 'connection failed', got '%s'", err.Error())
		}

		// Reset error for other tests
		testPlugin.SetConnectError(nil)
	})
}

func TestRPCValidate(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	t.Run("ValidConfig", func(t *testing.T) {
		config := map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"database": "testdb",
		}

		err := client.Validate(config)
		if err != nil {
			t.Errorf("Expected validation to pass, got error: %v", err)
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		testPlugin.SetValidateError(errors.New("missing required field: host"))

		err := client.Validate(map[string]interface{}{})
		if err == nil {
			t.Error("Expected validation error")
		}

		if err.Error() != "missing required field: host" {
			t.Errorf("Expected error message 'missing required field: host', got '%s'", err.Error())
		}

		// Reset error for other tests
		testPlugin.SetValidateError(nil)
	})
}

func TestRPCInitializeShutdown(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	t.Run("SuccessfulInitializeAndShutdown", func(t *testing.T) {
		// Test initialize
		ctx := context.Background()
		err := client.Initialize(ctx)
		if err != nil {
			t.Errorf("Expected initialize to succeed, got error: %v", err)
		}

		if !testPlugin.IsInitialized() {
			t.Error("Expected plugin to be initialized")
		}

		// Test shutdown
		err = client.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected shutdown to succeed, got error: %v", err)
		}

		if testPlugin.IsInitialized() {
			t.Error("Expected plugin to be shut down")
		}
	})

	t.Run("InitializeError", func(t *testing.T) {
		testPlugin.SetInitError(errors.New("initialization failed"))

		ctx := context.Background()
		err := client.Initialize(ctx)
		if err == nil {
			t.Error("Expected initialize error")
		}

		if err.Error() != "initialization failed" {
			t.Errorf("Expected error message 'initialization failed', got '%s'", err.Error())
		}

		// Reset error for other tests
		testPlugin.SetInitError(nil)
	})

	t.Run("ShutdownError", func(t *testing.T) {
		// First initialize successfully
		ctx := context.Background()
		err := client.Initialize(ctx)
		if err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}

		testPlugin.SetShutdownError(errors.New("shutdown failed"))

		err = client.Shutdown(ctx)
		if err == nil {
			t.Error("Expected shutdown error")
		}

		if err.Error() != "shutdown failed" {
			t.Errorf("Expected error message 'shutdown failed', got '%s'", err.Error())
		}

		// Reset error for other tests
		testPlugin.SetShutdownError(nil)
	})
}

func TestConnectionProxyMethods(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	// Create connection
	conn, err := client.Connect("testdb://localhost:5432/testdb", nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	proxy := conn.(*ConnectionProxy)

	t.Run("DatabaseName", func(t *testing.T) {
		name := proxy.DatabaseName()
		if name != "test_database" {
			t.Errorf("Expected database name 'test_database', got '%s'", name)
		}
	})

	t.Run("EngineVersion", func(t *testing.T) {
		version := proxy.EngineVersion()
		if version != "1.2.3" {
			t.Errorf("Expected engine version '1.2.3', got '%s'", version)
		}
	})

	t.Run("ListTables", func(t *testing.T) {
		tables, err := proxy.ListTables()
		if err != nil {
			t.Fatalf("Failed to list tables: %v", err)
		}

		expectedTables := []string{"users", "orders"}
		if len(tables) != len(expectedTables) {
			t.Errorf("Expected %d tables, got %d", len(expectedTables), len(tables))
		}

		for i, table := range tables {
			if table.Name != expectedTables[i] {
				t.Errorf("Expected table '%s', got '%s'", expectedTables[i], table.Name)
			}
			if table.Schema != "public" {
				t.Errorf("Expected schema 'public', got '%s'", table.Schema)
			}
		}
	})

	t.Run("ListTableForeignKeys", func(t *testing.T) {
		foreignKeys, err := proxy.ListTableForeignKeys("public", "orders")
		if err != nil {
			t.Fatalf("Failed to list foreign keys: %v", err)
		}

		if len(foreignKeys) != 1 {
			t.Errorf("Expected 1 foreign key, got %d", len(foreignKeys))
		}

		fk := foreignKeys[0]
		if fk.Name != "fk_orders_user_id" {
			t.Errorf("Expected foreign key name 'fk_orders_user_id', got '%s'", fk.Name)
		}
		if fk.ParentTable != "users" {
			t.Errorf("Expected parent table 'users', got '%s'", fk.ParentTable)
		}
		if len(fk.ChildColumns) != 1 || fk.ChildColumns[0] != "user_id" {
			t.Errorf("Expected child columns ['user_id'], got %v", fk.ChildColumns)
		}
	})

	t.Run("ListTableIndexes", func(t *testing.T) {
		indexes, err := proxy.ListTableIndexes("public", "users")
		if err != nil {
			t.Fatalf("Failed to list indexes: %v", err)
		}

		if len(indexes) != 2 {
			t.Errorf("Expected 2 indexes, got %d", len(indexes))
		}

		// Check first index
		idx := indexes[0]
		if idx.Name != "idx_users_email" {
			t.Errorf("Expected index name 'idx_users_email', got '%s'", idx.Name)
		}
		if !idx.IsUnique {
			t.Error("Expected index to be unique")
		}
	})

	t.Run("GetTablePrimaryKey", func(t *testing.T) {
		pk, err := proxy.GetTablePrimaryKey("users")
		if err != nil {
			t.Fatalf("Failed to get primary key: %v", err)
		}

		if pk.Name != "users_pkey" {
			t.Errorf("Expected primary key name 'users_pkey', got '%s'", pk.Name)
		}
		if !pk.IsPrimary {
			t.Error("Expected constraint to be primary")
		}
		if len(pk.Columns) != 1 || pk.Columns[0] != "id" {
			t.Errorf("Expected primary key columns ['id'], got %v", pk.Columns)
		}
	})

	t.Run("GetTableSchema", func(t *testing.T) {
		columns, err := proxy.GetTableSchema("users")
		if err != nil {
			t.Fatalf("Failed to get table schema: %v", err)
		}

		if len(columns) != 2 {
			t.Errorf("Expected 2 columns, got %d", len(columns))
		}

		// Check first column (id)
		idCol := columns[0]
		if idCol.Name != "id" {
			t.Errorf("Expected column name 'id', got '%s'", idCol.Name)
		}
		if idCol.DataType != "integer" {
			t.Errorf("Expected column type 'integer', got '%s'", idCol.DataType)
		}
		if idCol.Constraints == nil || idCol.Constraints.NotNull == nil || !*idCol.Constraints.NotNull {
			t.Error("Expected id column to be not null")
		}
		if idCol.Attributes == nil || idCol.Attributes.AutoIncrement == nil || !*idCol.Attributes.AutoIncrement {
			t.Error("Expected id column to be auto increment")
		}

		// Check second column (email)
		emailCol := columns[1]
		if emailCol.Name != "email" {
			t.Errorf("Expected column name 'email', got '%s'", emailCol.Name)
		}
		if emailCol.DataType != "varchar(255)" {
			t.Errorf("Expected column type 'varchar(255)', got '%s'", emailCol.DataType)
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := proxy.Close()
		if err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}

		// Verify connection is closed by checking if methods return errors
		_, err = proxy.ListTables()
		if err == nil {
			t.Error("Expected error when calling ListTables on closed connection")
		}
	})
}

func TestConnectionProxyErrorHandling(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, server, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	// Create connection first
	conn, err := client.Connect("testdb://localhost:5432/testdb", nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	proxy := conn.(*ConnectionProxy)

	t.Run("InvalidConnectionID", func(t *testing.T) {
		// Create a proxy with invalid connection ID
		invalidProxy := NewConnectionProxy(client.client, "invalid-connection-id")

		// Test that methods return appropriate errors
		name := invalidProxy.DatabaseName()
		if name != "" {
			t.Errorf("Expected empty database name for invalid connection, got '%s'", name)
		}

		version := invalidProxy.EngineVersion()
		if version != "" {
			t.Errorf("Expected empty engine version for invalid connection, got '%s'", version)
		}

		_, err := invalidProxy.ListTables()
		if err == nil {
			t.Error("Expected error for ListTables with invalid connection")
		}

		_, err = invalidProxy.ListTableForeignKeys("schema", "table")
		if err == nil {
			t.Error("Expected error for ListTableForeignKeys with invalid connection")
		}

		_, err = invalidProxy.ListTableIndexes("schema", "table")
		if err == nil {
			t.Error("Expected error for ListTableIndexes with invalid connection")
		}

		_, err = invalidProxy.GetTablePrimaryKey("table")
		if err == nil {
			t.Error("Expected error for GetTablePrimaryKey with invalid connection")
		}

		_, err = invalidProxy.GetTableSchema("table")
		if err == nil {
			t.Error("Expected error for GetTableSchema with invalid connection")
		}

		err = invalidProxy.Close()
		if err == nil {
			t.Error("Expected error for Close with invalid connection")
		}
	})

	t.Run("DoubleClose", func(t *testing.T) {
		// First close should succeed
		err := proxy.Close()
		if err != nil {
			t.Errorf("First close should succeed: %v", err)
		}

		// Second close should not fail (connection already removed from server)
		err = proxy.Close()
		if err == nil {
			t.Error("Expected error on second close")
		}
	})

	_ = server // Keep server reference to avoid unused variable
}

func TestDataSerialization(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	// Create connection
	conn, err := client.Connect("testdb://localhost:5432/testdb", nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	proxy := conn.(*ConnectionProxy)

	t.Run("ComplexDataTypes", func(t *testing.T) {
		// Test that complex data structures are properly serialized/deserialized
		tables, err := proxy.ListTables()
		if err != nil {
			t.Fatalf("Failed to list tables: %v", err)
		}

		// Verify all fields are preserved
		for _, table := range tables {
			if table.Name == "" {
				t.Error("Table name should not be empty")
			}
			if table.Schema == "" {
				t.Error("Table schema should not be empty")
			}
			if table.Charset == "" {
				t.Error("Table charset should not be empty")
			}
			if table.Collation == "" {
				t.Error("Table collation should not be empty")
			}
		}
	})

	t.Run("NilPointers", func(t *testing.T) {
		// Test handling of nil pointers in complex structures
		columns, err := proxy.GetTableSchema("users")
		if err != nil {
			t.Fatalf("Failed to get table schema: %v", err)
		}

		// Verify pointer fields are properly handled
		for _, col := range columns {
			if col.Constraints != nil && col.Constraints.NotNull != nil {
				// This should not panic and should preserve the boolean value
				notNull := *col.Constraints.NotNull
				if col.Name == "id" && !notNull {
					t.Error("Expected id column to be not null")
				}
			}
		}
	})

	t.Run("MapSerialization", func(t *testing.T) {
		// Test that map data is properly serialized/deserialized
		// Note: RPC has limitations on complex nested types, so we test with simpler structures
		options := map[string]interface{}{
			"timeout": 30,
			"ssl":     true,
			"host":    "localhost",
			"port":    5432,
		}

		// Connect with options
		conn2, err := client.Connect("testdb://localhost:5432/testdb2", options)
		if err != nil {
			t.Fatalf("Failed to connect with options: %v", err)
		}

		// If we get here, the options were successfully serialized/deserialized
		proxy2 := conn2.(*ConnectionProxy)
		if proxy2.connectionID == "" {
			t.Error("Expected non-empty connection ID")
		}

		// Clean up
		err = proxy2.Close()
		if err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	})
}

func TestErrorPropagation(t *testing.T) {
	testPlugin := NewTestPlugin()
	client, _, err := createInProcessRPCClientServer(testPlugin)
	if err != nil {
		t.Fatalf("Failed to create RPC client/server: %v", err)
	}
	defer client.client.Close()

	t.Run("PluginErrors", func(t *testing.T) {
		// Test that errors from the plugin implementation are properly propagated
		originalError := fmt.Errorf("custom plugin error with details: %s", "some context")
		testPlugin.SetValidateError(originalError)

		err := client.Validate(map[string]interface{}{})
		if err == nil {
			t.Fatal("Expected error to be propagated")
		}

		expectedMsg := "custom plugin error with details: some context"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}

		// Reset for other tests
		testPlugin.SetValidateError(nil)
	})

	t.Run("ConnectionErrors", func(t *testing.T) {
		// Create connection and close it to simulate closed connection errors
		conn, err := client.Connect("testdb://localhost:5432/testdb", nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		proxy := conn.(*ConnectionProxy)
		err = proxy.Close()
		if err != nil {
			t.Fatalf("Failed to close connection: %v", err)
		}

		// Now try to use closed connection - should get appropriate errors
		_, err = proxy.ListTables()
		if err == nil {
			t.Error("Expected error when using closed connection")
		}

		// The connection is removed from the server when closed, so we get "connection not found"
		expectedError := fmt.Sprintf("connection %s not found", proxy.connectionID)
		if err.Error() != expectedError {
			t.Errorf("Expected '%s' error, got '%s'", expectedError, err.Error())
		}
	})
}
