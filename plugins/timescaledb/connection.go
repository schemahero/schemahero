package main

import (
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
	postgres "github.com/schemahero/schemahero/plugins/postgres/lib"
	timescaledb "github.com/schemahero/schemahero/plugins/timescaledb/internal"
)

// TimescaleDBConnection wraps a PostgreSQL connection and adds TimescaleDB-specific functionality
type TimescaleDBConnection struct {
	*postgres.PostgresConnection
	uri string
}

// NewTimescaleDBConnection creates a new TimescaleDB connection wrapper
func NewTimescaleDBConnection(pgConn *postgres.PostgresConnection, uri string) *TimescaleDBConnection {
	return &TimescaleDBConnection{
		PostgresConnection: pgConn,
		uri:               uri,
	}
}

// NewFixtureOnlyTimescaleDBConnection creates a TimescaleDB connection that only supports fixture generation
// without actually connecting to a database. This is used for fixture DDL generation.
func NewFixtureOnlyTimescaleDBConnection() *TimescaleDBConnection {
	// Use the fixture-only postgres connection
	fixtureOnlyPgConn := postgres.NewFixtureOnlyConnection()
	
	return &TimescaleDBConnection{
		PostgresConnection: fixtureOnlyPgConn,
		uri:               "",
	}
}

// PlanTableSchema generates SQL statements to migrate a table to the desired schema
// This overrides the PostgreSQL implementation to handle TimescaleDB-specific features
func (t *TimescaleDBConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	tsSchema, ok := tableSchema.(*schemasv1alpha4.TimescaleDBTableSchema)
	if !ok {
		// If it's not a TimescaleDB schema, fall back to PostgreSQL
		return t.PostgresConnection.PlanTableSchema(tableName, tableSchema, seedData)
	}
	
	// Use TimescaleDB-specific planning
	return timescaledb.PlanTimescaleDBTable(t.uri, tableName, tsSchema, seedData)
}

// The following methods delegate to the embedded PostgresConnection
func (t *TimescaleDBConnection) Close() error {
	return t.PostgresConnection.Close()
}

func (t *TimescaleDBConnection) DatabaseName() string {
	return t.PostgresConnection.DatabaseName()
}

func (t *TimescaleDBConnection) EngineVersion() string {
	return t.PostgresConnection.EngineVersion()
}

func (t *TimescaleDBConnection) ListTables() ([]*types.Table, error) {
	return t.PostgresConnection.ListTables()
}

func (t *TimescaleDBConnection) ListTableForeignKeys(databaseName, tableName string) ([]*types.ForeignKey, error) {
	return t.PostgresConnection.ListTableForeignKeys(databaseName, tableName)
}

func (t *TimescaleDBConnection) ListTableIndexes(databaseName, tableName string) ([]*types.Index, error) {
	return t.PostgresConnection.ListTableIndexes(databaseName, tableName)
}

func (t *TimescaleDBConnection) GetTablePrimaryKey(tableName string) (*types.KeyConstraint, error) {
	return t.PostgresConnection.GetTablePrimaryKey(tableName)
}

func (t *TimescaleDBConnection) GetTableSchema(tableName string) ([]*types.Column, error) {
	return t.PostgresConnection.GetTableSchema(tableName)
}

func (t *TimescaleDBConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	// Check if it's a TimescaleDB view schema
	tsViewSchema, ok := viewSchema.(*schemasv1alpha4.TimescaleDBViewSchema)
	if ok {
		// Use TimescaleDB-specific view planning
		return timescaledb.PlanTimescaleDBView(t.uri, viewName, tsViewSchema)
	}
	
	// Fall back to PostgreSQL for regular views
	return t.PostgresConnection.PlanViewSchema(viewName, viewSchema)
}

func (t *TimescaleDBConnection) PlanTypeSchema(typeName string, typeSchema interface{}) ([]string, error) {
	// Delegate to PostgreSQL connection for type schemas
	return t.PostgresConnection.PlanTypeSchema(typeName, typeSchema)
}

func (t *TimescaleDBConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	return t.PostgresConnection.PlanFunctionSchema(functionName, functionSchema)
}

func (t *TimescaleDBConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	return t.PostgresConnection.PlanExtensionSchema(extensionName, extensionSchema)
}

func (t *TimescaleDBConnection) DeployStatements(statements []string) error {
	return t.PostgresConnection.DeployStatements(statements)
}

// GenerateFixtures generates SQL statements to create tables and seed data for fixtures
func (t *TimescaleDBConnection) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil {
		return []string{}, nil
	}

	// Check if it's a TimescaleDB schema first
	if spec.Schema.TimescaleDB != nil {
		// Skip deleted tables
		if spec.Schema.TimescaleDB.IsDeleted {
			return []string{}, nil
		}

		// Generate TimescaleDB table statements
		statements, err := timescaledb.CreateTableStatements(spec.Name, spec.Schema.TimescaleDB)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statements")
		}

		// Add seed data if present
		if spec.SeedData != nil {
			seedStatements, err := timescaledb.SeedDataStatements(spec.Name, spec.Schema.TimescaleDB, spec.SeedData)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create seed data statements")
			}
			statements = append(statements, seedStatements...)
		}

		return statements, nil
	}

	// Fall back to PostgreSQL if not a TimescaleDB schema
	return t.PostgresConnection.GenerateFixtures(spec)
}