package cassandra

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

type CassandraConnection struct {
	session  *gocql.Session
	keyspace string
}

func Connect(hosts []string, username string, password string, keyspace string) (*CassandraConnection, error) {
	cluster := gocql.NewCluster(hosts...)

	// Set the keyspace if provided
	if keyspace != "" {
		cluster.Keyspace = keyspace
	}

	if username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: username,
			Password: password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cassandra session")
	}

	cassandraConnection := CassandraConnection{
		session:  session,
		keyspace: keyspace,
	}

	return &cassandraConnection, nil
}

func (c *CassandraConnection) Close() error {
	if c.session == nil {
		return nil
	}
	c.session.Close()
	return nil
}

// IsConnected checks if the connection is still active
func (c *CassandraConnection) IsConnected() bool {
	if c.session == nil {
		return false
	}
	// Check if we can execute a simple query
	if err := c.session.Query("SELECT now() FROM system.local").Exec(); err != nil {
		return false
	}
	return true
}

// DatabaseName returns the keyspace name
func (c *CassandraConnection) DatabaseName() string {
	return c.keyspace
}

// EngineVersion returns the Cassandra version
func (c *CassandraConnection) EngineVersion() string {
	if c.session == nil {
		return "fixture-only"
	}
	var version string
	if err := c.session.Query("SELECT release_version FROM system.local").Scan(&version); err != nil {
		return "unknown"
	}
	return version
}

// ListTables returns all tables in the keyspace
func (c *CassandraConnection) ListTables() ([]*types.Table, error) {
	iter := c.session.Query("SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?", c.keyspace).Iter()
	tables := []*types.Table{}
	var tableName string
	for iter.Scan(&tableName) {
		tables = append(tables, &types.Table{
			Name: tableName,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}
	return tables, nil
}

// ListTableForeignKeys - Cassandra doesn't support foreign keys
func (c *CassandraConnection) ListTableForeignKeys(databaseName, tableName string) ([]*types.ForeignKey, error) {
	// Cassandra doesn't support foreign keys
	return []*types.ForeignKey{}, nil
}

// ListTableIndexes returns all indexes for a table
func (c *CassandraConnection) ListTableIndexes(databaseName, tableName string) ([]*types.Index, error) {
	iter := c.session.Query("SELECT index_name, options FROM system_schema.indexes WHERE keyspace_name = ? AND table_name = ?", c.keyspace, tableName).Iter()
	indexes := []*types.Index{}
	var indexName string
	var options map[string]string
	for iter.Scan(&indexName, &options) {
		index := &types.Index{
			Name: indexName,
		}
		// Extract target column from options if available
		if target, exists := options["target"]; exists {
			index.Columns = []string{target}
		}
		indexes = append(indexes, index)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to list indexes")
	}
	return indexes, nil
}

// GetTablePrimaryKey returns the primary key constraint for a table
func (c *CassandraConnection) GetTablePrimaryKey(tableName string) (*types.KeyConstraint, error) {
	// Query the partition key and clustering columns
	var partitionKeys []string
	var clusteringKeys []string
	
	// Get partition key columns
	iter := c.session.Query("SELECT column_name FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ? AND kind = 'partition_key' ALLOW FILTERING", c.keyspace, tableName).Iter()
	var columnName string
	for iter.Scan(&columnName) {
		partitionKeys = append(partitionKeys, columnName)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to get partition keys")
	}
	
	// Get clustering key columns
	iter = c.session.Query("SELECT column_name FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ? AND kind = 'clustering' ALLOW FILTERING", c.keyspace, tableName).Iter()
	for iter.Scan(&columnName) {
		clusteringKeys = append(clusteringKeys, columnName)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to get clustering keys")
	}
	
	// Combine partition and clustering keys
	allKeys := append(partitionKeys, clusteringKeys...)
	if len(allKeys) == 0 {
		return nil, nil
	}
	
	return &types.KeyConstraint{
		Columns: allKeys,
	}, nil
}

// GetTableSchema returns all columns for a table
func (c *CassandraConnection) GetTableSchema(tableName string) ([]*types.Column, error) {
	iter := c.session.Query("SELECT column_name, type, kind FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?", c.keyspace, tableName).Iter()
	columns := []*types.Column{}
	var columnName, columnType, kind string
	for iter.Scan(&columnName, &columnType, &kind) {
		column := &types.Column{
			Name:     columnName,
			DataType: columnType,
		}
		// Set primary key flag for partition and clustering keys
		if kind == "partition_key" || kind == "clustering" {
			column.Constraints = &types.ColumnConstraints{
				NotNull: &[]bool{true}[0],
			}
		}
		columns = append(columns, column)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to get table schema")
	}
	return columns, nil
}

// PlanTableSchema generates SQL statements for table schema changes
func (c *CassandraConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Type assert to CassandraTableSchema
	schema, ok := tableSchema.(*schemasv1alpha4.CassandraTableSchema)
	if !ok {
		return nil, fmt.Errorf("expected CassandraTableSchema, got %T", tableSchema)
	}
	// Use the connection-based planning function
	return c.planCassandraTable(tableName, schema, seedData)
}

// PlanTypeSchema generates SQL statements for type (UDT) schema changes
func (c *CassandraConnection) PlanTypeSchema(typeName string, typeSchema interface{}) ([]string, error) {
	// Type assert to CassandraDataTypeSchema
	schema, ok := typeSchema.(*schemasv1alpha4.CassandraDataTypeSchema)
	if !ok {
		return nil, fmt.Errorf("expected CassandraDataTypeSchema, got %T", typeSchema)
	}
	
	// Check if the type exists
	query := `select count(1) from system_schema.types where keyspace_name=? and type_name = ?`
	row := c.session.Query(query, c.keyspace, typeName)
	typeExists := 0
	if err := row.Scan(&typeExists); err != nil {
		return nil, errors.Wrap(err, "failed to check if type exists")
	}
	
	if typeExists == 0 && schema.IsDeleted {
		return []string{}, nil
	} else if typeExists > 0 && schema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop type "%s"`, typeName),
		}, nil
	}
	
	if typeExists == 0 {
		// Create the type
		fields := []string{}
		for _, field := range schema.Fields {
			fields = append(fields, fmt.Sprintf("%s %s", field.Name, field.Type))
		}
		
		return []string{
			fmt.Sprintf(`create type "%s" (%s)`, typeName, strings.Join(fields, ", ")),
		}, nil
	}
	
	// Cassandra doesn't support altering types
	// You can only drop and recreate them
	// For now, return empty if type exists and matches
	return []string{}, nil
}

// PlanViewSchema - Cassandra views are not yet implemented
func (c *CassandraConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	return nil, errors.New("cassandra view planning not yet implemented")
}

// PlanFunctionSchema - Cassandra doesn't support stored functions
func (c *CassandraConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	return nil, errors.New("cassandra does not support stored functions")
}

// PlanExtensionSchema - Cassandra doesn't support extensions
func (c *CassandraConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	return nil, errors.New("cassandra does not support extensions")
}

// DeployStatements executes the provided SQL statements
func (c *CassandraConnection) DeployStatements(statements []string) error {
	// Execute statements directly using the connection
	for _, statement := range statements {
		if statement == "" {
			continue
		}
		// Statement is already printed by the main process
		if err := c.session.Query(statement).Exec(); err != nil {
			return errors.Wrapf(err, "failed to execute statement: %s", statement)
		}
	}
	return nil
}

// planCassandraTable is the connection-based planning function
func (c *CassandraConnection) planCassandraTable(tableName string, cassandraTableSchema *schemasv1alpha4.CassandraTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// determine if the table exists
	query := `select count(1) from system_schema.tables where keyspace_name=? and table_name = ?`
	row := c.session.Query(query, c.keyspace, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && cassandraTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && cassandraTableSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop table %s.%s`, c.keyspace, tableName),
		}, nil
	}

	if tableExists == 0 {
		// shortcut to just create it
		queries, err := CreateTableStatements(c.keyspace, tableName, cassandraTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return queries, nil
	}

	statements := []string{}

	columnStatements, err := buildColumnStatements(c, tableName, cassandraTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, columnStatements...)

	propertiesStatements, err := buildPropertiesStatements(c, tableName, cassandraTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build properties statements")
	}
	statements = append(statements, propertiesStatements...)

	return statements, nil
}

// GenerateFixtures generates SQL statements to create tables and seed data for fixtures
func (c *CassandraConnection) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil || spec.Schema.Cassandra == nil {
		return []string{}, nil
	}

	// Skip deleted tables
	if spec.Schema.Cassandra.IsDeleted {
		return []string{}, nil
	}

	// Generate create table statements
	statements, err := CreateTableStatements(c.keyspace, spec.Name, spec.Schema.Cassandra)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create table statements")
	}

	// Add seed data if present
	if spec.SeedData != nil {
		seedStatements, err := SeedDataStatements(c.keyspace, spec.Name, spec.SeedData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create seed data statements")
		}
		statements = append(statements, seedStatements...)
	}

	return statements, nil
}

// NewFixtureOnlyConnection creates a Cassandra connection that only supports fixture generation
// without actually connecting to a database. This is used for fixture DDL generation.
func NewFixtureOnlyConnection() *CassandraConnection {
	return &CassandraConnection{
		session:  nil, // No actual database connection
		keyspace: "schemahero",
	}
}
