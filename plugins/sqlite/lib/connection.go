package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type SqliteConnection struct {
	db  *sql.DB
	uri string
}

func Connect(dsn string) (*SqliteConnection, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	sqliteConnection := SqliteConnection{
		db:  db,
		uri: dsn,
	}

	return &sqliteConnection, nil
}

func (s *SqliteConnection) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// IsConnected returns true if the connection is alive
func (s *SqliteConnection) IsConnected() bool {
	if s.db == nil {
		return false
	}
	err := s.db.Ping()
	return err == nil
}

// DatabaseName returns the database name (file path for SQLite)
func (s *SqliteConnection) DatabaseName() string {
	// SQLite uses file paths as database names
	return s.uri
}

// EngineVersion returns the SQLite version
func (s *SqliteConnection) EngineVersion() string {
	if s.db == nil {
		return "fixture-only"
	}
	var version string
	row := s.db.QueryRow("SELECT sqlite_version()")
	if err := row.Scan(&version); err != nil {
		return "unknown"
	}
	return version
}

// PlanTypeSchema generates SQL statements for type schema (not supported by SQLite)
func (s *SqliteConnection) PlanTypeSchema(typeName string, typeSchema interface{}) ([]string, error) {
	return nil, errors.New("type schemas are not supported in SQLite")
}

// PlanTableSchema generates SQL statements to reconcile a table schema
func (s *SqliteConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Handle seed data without schema case
	if tableSchema == nil && seedData != nil {
		// Need to verify the table exists and generate seed data statements
		return PlanSqliteTableSeedDataOnly(s.uri, tableName, seedData)
	}

	sqliteTableSchema, ok := tableSchema.(*schemasv1alpha4.SqliteTableSchema)
	if !ok {
		return nil, fmt.Errorf("expected SqliteTableSchema, got %T", tableSchema)
	}
	return PlanSqliteTable(s.uri, tableName, sqliteTableSchema, seedData)
}

// PlanViewSchema generates SQL statements for managing views
func (s *SqliteConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	// SQLite view support not yet implemented
	return nil, errors.New("SQLite view planning not yet implemented")
}

// PlanFunctionSchema generates SQL statements for managing functions
func (s *SqliteConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	// SQLite doesn't support stored functions
	return nil, errors.New("SQLite does not support stored functions")
}

// PlanExtensionSchema generates SQL statements for managing extensions
func (s *SqliteConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	// SQLite doesn't support extensions
	return nil, errors.New("SQLite does not support extensions")
}

// DeployStatements executes a list of SQL statements
func (s *SqliteConnection) DeployStatements(statements []string) error {
	return DeploySqliteStatements(s.uri, statements)
}

// GenerateFixtures generates SQL statements to create tables and seed data for fixtures
func (s *SqliteConnection) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil || spec.Schema.SQLite == nil {
		return []string{}, nil
	}

	// Skip deleted tables
	if spec.Schema.SQLite.IsDeleted {
		return []string{}, nil
	}

	// Generate create table statements
	statements, err := CreateTableStatements(spec.Name, spec.Schema.SQLite)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create table statements")
	}

	// Add seed data if present
	if spec.SeedData != nil {
		seedStatements, err := SeedDataStatements(spec.Name, spec.SeedData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create seed data statements")
		}
		statements = append(statements, seedStatements...)
	}

	return statements, nil
}

// GetDB returns the underlying database connection
func (s *SqliteConnection) GetDB() *sql.DB {
	return s.db
}

// NewFixtureOnlyConnection creates a SQLite connection that only supports fixture generation
// without actually connecting to a database. This is used for fixture DDL generation.
func NewFixtureOnlyConnection() *SqliteConnection {
	return &SqliteConnection{
		db:  nil, // No actual database connection
		uri: ":memory:",
	}
}
