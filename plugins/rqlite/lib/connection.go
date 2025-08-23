package rqlite

import (
	"net"
	nurl "net/url"

	"github.com/pkg/errors"
	"github.com/rqlite/gorqlite"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type RqliteConnection struct {
	db *gorqlite.Connection
}

func Connect(url string) (*RqliteConnection, error) {
	db, err := gorqlite.Open(url)
	if err != nil {
		return nil, err
	}

	rqliteConnection := RqliteConnection{
		db: db,
	}

	return &rqliteConnection, nil
}

func (s RqliteConnection) Close() error {
	s.db.Close()
	return nil
}

func (m *RqliteConnection) DatabaseName() string {
	return ""
}

func (p *RqliteConnection) EngineVersion() string {
	return ""
}

// PlanTableSchema implements interfaces.SchemaHeroDatabaseConnection.PlanTableSchema()
func (r *RqliteConnection) PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// RQLite planning not yet implemented
	return nil, errors.New("RQLite table planning not yet implemented")
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
	// RQLite deployment not yet implemented
	return errors.New("RQLite statement deployment not yet implemented")
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
