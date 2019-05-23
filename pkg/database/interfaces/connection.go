package interfaces

import "database/sql"

type SchemaHeroDatabaseConnection interface {
	GetConnection() *sql.Conn
	GetDB() *sql.DB

	DatabaseName() string
	EngineVersion() string

	CheckAlive(string, string) (bool, error)
}
