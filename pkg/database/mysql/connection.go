package mysql

import (
	"context"
	"database/sql"
	"strings"

	// import the mysql driver
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/trace"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type MysqlConnection struct {
	db            *sql.DB
	databaseName  string
	engineVersion string
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

func Connect(ctx context.Context, uri string) (*MysqlConnection, error) {
	var span oteltrace.Span
	ctx, span = otel.Tracer(trace.TraceName).Start(ctx, "Connect")
	defer span.End()

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
	}

	return &mysqlConnection, nil
}

func (m *MysqlConnection) Close() error {
	if m.db == nil {
		return nil
	}
	return errors.Wrap(m.db.Close(), "failed to close connection")
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
