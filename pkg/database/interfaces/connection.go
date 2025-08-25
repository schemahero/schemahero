package interfaces

import (
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

type SchemaHeroDatabaseConnection interface {
	Close() error

	DatabaseName() string
	EngineVersion() string

	ListTables() ([]*types.Table, error)
	ListTableForeignKeys(string, string) ([]*types.ForeignKey, error)
	ListTableIndexes(string, string) ([]*types.Index, error)

	GetTablePrimaryKey(string) (*types.KeyConstraint, error)
	GetTableSchema(string) ([]*types.Column, error)

	// Planning methods - generate SQL statements for schema changes
	PlanTableSchema(tableName string, tableSchema interface{}, seedData *schemasv1alpha4.SeedData) ([]string, error)
	PlanViewSchema(viewName string, viewSchema interface{}) ([]string, error)
	PlanFunctionSchema(functionName string, functionSchema interface{}) ([]string, error)
	PlanExtensionSchema(extensionName string, extensionSchema interface{}) ([]string, error)

	// Deployment methods - execute SQL statements
	DeployStatements(statements []string) error

	// Fixture generation
	GenerateFixtures(spec *schemasv1alpha4.TableSpec) ([]string, error)
}
