package rqlite

import (
	"net"
	nurl "net/url"

	"github.com/pkg/errors"
	"github.com/rqlite/gorqlite"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type RqliteConnection struct {
	db  *gorqlite.Connection
	uri string
}

func Connect(url string) (*RqliteConnection, error) {
	db, err := gorqlite.Open(url)
	if err != nil {
		return nil, err
	}

	rqliteConnection := RqliteConnection{
		db:  db,
		uri: url,
	}

	return &rqliteConnection, nil
}

func (s RqliteConnection) Close() error {
	if s.db == nil {
		return nil
	}
	s.db.Close()
	return nil
}

func (m *RqliteConnection) DatabaseName() string {
	return ""
}

func (p *RqliteConnection) EngineVersion() string {
	if p.db == nil {
		return "fixture-only"
	}
	return ""
}

// PlanTableSchema implements interfaces.SchemaHeroDatabaseConnection.PlanTableSchema()
func (r *RqliteConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Type assert to the correct schema type
	rqliteSchema, ok := tableSchema.(*schemasv1alpha4.RqliteTableSchema)
	if !ok {
		return nil, errors.New("tableSchema must be *RqliteTableSchema")
	}
	
	// Use the existing RQLite planning implementation
	// Note: PlanRqliteTable expects a URL, but we need to pass the URL from somewhere
	// For now, we'll need to store it in the connection
	if r.uri == "" {
		return nil, errors.New("URI not set in RqliteConnection")
	}
	return PlanRqliteTable(r.uri, tableName, rqliteSchema, seedData)
}

// PlanViewSchema implements interfaces.SchemaHeroDatabaseConnection.PlanViewSchema()
func (r *RqliteConnection) PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error) {
	// RQLite doesn't support views
	return nil, errors.New("RQLite does not support views")
}

// PlanFunctionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanFunctionSchema()
func (r *RqliteConnection) PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error) {
	// RQLite doesn't support stored functions
	return nil, errors.New("RQLite does not support stored functions")
}

// PlanExtensionSchema implements interfaces.SchemaHeroDatabaseConnection.PlanExtensionSchema()
func (r *RqliteConnection) PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error) {
	// RQLite doesn't support extensions
	return nil, errors.New("RQLite does not support extensions")
}

// DeployStatements implements interfaces.SchemaHeroDatabaseConnection.DeployStatements()
func (r *RqliteConnection) DeployStatements(statements []string) error {
	if r.uri == "" {
		return errors.New("URI not set in RqliteConnection")
	}
	return DeployRqliteStatements(r.uri, statements)
}

// GenerateFixtures generates SQL statements to create tables and seed data for fixtures
func (r *RqliteConnection) GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil || spec.Schema.RQLite == nil {
		return []string{}, nil
	}

	// Skip deleted tables
	if spec.Schema.RQLite.IsDeleted {
		return []string{}, nil
	}

	// Generate create table statements
	statements, err := CreateTableStatements(spec.Name, spec.Schema.RQLite)
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

// NewFixtureOnlyConnection creates a RQLite connection that only supports fixture generation
// without actually connecting to a database. This is used for fixture DDL generation.
func NewFixtureOnlyConnection() *RqliteConnection {
	return &RqliteConnection{
		db: nil, // No actual database connection
	}
}

func UsernameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	return u.User.Username(), nil
}

func PasswordFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	pass, _ := u.User.Password()
	return pass, nil
}

func HostnameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return host, nil
}

func PortFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return port, nil
}
