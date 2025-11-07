package mysql

import (
	"database/sql"
	"strings"

	// import the mysql driver
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type MysqlConnection struct {
	db            *sql.DB
	databaseName  string
	engineVersion string
	uri           string // Store the connection URI
}

func (m *MysqlConnection) GetDB() *sql.DB {
	return m.db
}

func (m *MysqlConnection) DatabaseName() string {
	return m.databaseName
}

func (m *MysqlConnection) EngineVersion() string {
	return m.engineVersion
}

func Connect(uri string) (*MysqlConnection, error) {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	databaseName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return nil, err
	}

	mysqlConnection := MysqlConnection{
		db:           db,
		databaseName: databaseName,
		uri:          uri,
	}

	return &mysqlConnection, nil
}

func (m *MysqlConnection) Close() error {
	if m.db == nil {
		return nil
	}
	return errors.Wrap(m.db.Close(), "failed to close connection")
}

// PlanTypeSchema generates SQL statements for type schema (not supported by MySQL)
func (m *MysqlConnection) PlanTypeSchema(typeName string, typeSchema interface{}) ([]string, error) {
	return nil, errors.New("type schemas are not supported in MySQL")
}

// PlanTableSchema generates SQL statements to migrate a table to the desired schema
func (m *MysqlConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Handle seed data without schema case
	if tableSchema == nil && seedData != nil {
		// Need to retrieve the existing table schema from the database
		// and generate seed data statements based on that
		return PlanMysqlTableSeedDataOnly(m.uri, tableName, seedData)
	}

	mysqlSchema, ok := tableSchema.(*schemasv1alpha4.MysqlTableSchema)
	if !ok {
		return nil, errors.New("tableSchema must be *MysqlTableSchema")
	}

	return PlanMysqlTable(m.uri, tableName, mysqlSchema, seedData)
}

// PlanViewSchema generates SQL statements to create or update a view
func (m *MysqlConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	mysqlView, ok := viewSchema.(*schemasv1alpha4.NotImplementedViewSchema)
	if !ok {
		return nil, errors.New("viewSchema must be *NotImplementedViewSchema")
	}
	
	return PlanMysqlView(m.uri, viewName, mysqlView)
}

// PlanFunctionSchema generates SQL statements to create or update a function
func (m *MysqlConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	// MySQL functions not yet implemented
	return nil, errors.New("MySQL function planning not yet implemented")
}

// PlanExtensionSchema generates SQL statements to manage database extensions
func (m *MysqlConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	// MySQL doesn't have extensions like PostgreSQL
	return nil, errors.New("MySQL does not support extensions")
}

// DeployStatements executes a list of SQL statements
func (m *MysqlConnection) DeployStatements(statements []string) error {
	return DeployMysqlStatements(m.uri, statements)
}

func (m *MysqlConnection) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil || spec.Schema.Mysql == nil {
		return []string{}, nil
	}

	// Skip deleted tables
	if spec.Schema.Mysql.IsDeleted {
		return []string{}, nil
	}

	// Generate create table statements
	statements, err := CreateTableStatements(spec.Name, spec.Schema.Mysql)
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

// NewFixtureOnlyConnection creates a MySQL connection that only supports fixture generation
// without actually connecting to a database. This is used for fixture DDL generation.
func NewFixtureOnlyConnection() *MysqlConnection {
	return &MysqlConnection{
		db:            nil,  // No actual database connection
		databaseName:  "schemahero",
		engineVersion: "fixture-only",
		uri:           "",
	}
}

func DatabaseNameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.DBName, nil
}

func UsernameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.User, nil
}

func PasswordFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.Passwd, nil
}

func HostnameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	// TODO this is very happy path
	addrAndPort := strings.Split(cfg.Addr, ":")
	return addrAndPort[0], nil
}

func PortFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	// TODO this is very happy path
	addrAndPort := strings.Split(cfg.Addr, ":")

	if len(addrAndPort) > 1 {
		return addrAndPort[1], nil
	}
	return "3306", nil
}
