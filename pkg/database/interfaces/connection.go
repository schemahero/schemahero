package interfaces

import (
	"database/sql"

	"github.com/schemahero/schemahero/pkg/database/types"
)

type SchemaHeroDatabaseConnection interface {
	GetConnection() *sql.Conn
	GetDB() *sql.DB

	DatabaseName() string
	EngineVersion() string

	CheckAlive(string, string) (bool, error)

	ListTables() ([]string, error)
	ListTableForeignKeys(string, string) ([]*types.ForeignKey, error)
	ListTableIndexes(string, string) ([]*types.Index, error)

	GetTablePrimaryKey(string) (*types.KeyConstraint, error)
	GetTableSchema(string) ([]*types.Column, error)
}
