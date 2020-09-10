package interfaces

import (
	"github.com/schemahero/schemahero/pkg/database/types"
)

type SchemaHeroDatabaseConnection interface {
	Close()

	DatabaseName() string
	EngineVersion() string

	ListTables() ([]*types.Table, error)
	ListTableForeignKeys(string, string) ([]*types.ForeignKey, error)
	ListTableIndexes(string, string) ([]*types.Index, error)

	GetTablePrimaryKey(string) (*types.KeyConstraint, error)
	GetTableSchema(string) ([]*types.Column, error)
}
