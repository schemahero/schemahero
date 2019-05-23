package mysql

import (
	"database/sql"

	// import the mysql driver
	mysqldriver "github.com/go-sql-driver/mysql"
)

type MysqlConnection struct {
	conn          *sql.Conn
	db            *sql.DB
	databaseName  string
	engineVersion string
}

func (m *MysqlConnection) GetConnection() *sql.Conn {
	return m.conn
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
	}

	return &mysqlConnection, nil
}

func DatabaseNameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", err
	}

	return cfg.DBName, nil
}
