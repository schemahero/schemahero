package types

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type ForeignKey struct {
	ChildColumns  []string
	ParentTable   string
	ParentColumns []string
	Name          string
	OnDelete      string
}

func (fk *ForeignKey) Equals(other *ForeignKey) bool {
	// TODO

	return false
}

func ForeignKeyToMysqlSchemaForeignKey(foreignKey *ForeignKey) *schemasv1alpha4.MysqlTableForeignKey {
	schemaForeignKey := schemasv1alpha4.MysqlTableForeignKey{
		Columns: foreignKey.ChildColumns,
		References: schemasv1alpha4.MysqlTableForeignKeyReferences{
			Table:   foreignKey.ParentTable,
			Columns: foreignKey.ParentColumns,
		},
		Name:     foreignKey.Name,
		OnDelete: foreignKey.OnDelete,
	}

	return &schemaForeignKey
}

func ForeignKeyToPostgresqlSchemaForeignKey(foreignKey *ForeignKey) *schemasv1alpha4.PostgresqlTableForeignKey {
	schemaForeignKey := schemasv1alpha4.PostgresqlTableForeignKey{
		Columns: foreignKey.ChildColumns,
		References: schemasv1alpha4.PostgresqlTableForeignKeyReferences{
			Table:   foreignKey.ParentTable,
			Columns: foreignKey.ParentColumns,
		},
		Name:     foreignKey.Name,
		OnDelete: foreignKey.OnDelete,
	}

	return &schemaForeignKey
}

func MysqlSchemaForeignKeyToForeignKey(schemaForeignKey *schemasv1alpha4.MysqlTableForeignKey) *ForeignKey {
	foreignKey := ForeignKey{
		ChildColumns:  schemaForeignKey.Columns,
		ParentTable:   schemaForeignKey.References.Table,
		ParentColumns: schemaForeignKey.References.Columns,
		Name:          schemaForeignKey.Name,
		OnDelete:      schemaForeignKey.OnDelete,
	}

	return &foreignKey
}

func PostgresqlSchemaForeignKeyToForeignKey(schemaForeignKey *schemasv1alpha4.PostgresqlTableForeignKey) *ForeignKey {
	foreignKey := ForeignKey{
		ChildColumns:  schemaForeignKey.Columns,
		ParentTable:   schemaForeignKey.References.Table,
		ParentColumns: schemaForeignKey.References.Columns,
		Name:          schemaForeignKey.Name,
		OnDelete:      schemaForeignKey.OnDelete,
	}

	return &foreignKey
}

func GenerateMysqlFKName(tableName string, schemaForeignKey *schemasv1alpha4.MysqlTableForeignKey) string {
	if schemaForeignKey.Name != "" {
		return schemaForeignKey.Name
	}

	return fmt.Sprintf("%s_%s_fkey", tableName, strings.Join(schemaForeignKey.Columns, "_"))
}

func GeneratePostgresqlFKName(tableName string, schemaForeignKey *schemasv1alpha4.PostgresqlTableForeignKey) string {
	if schemaForeignKey.Name != "" {
		return schemaForeignKey.Name
	}

	return fmt.Sprintf("%s_%s_fkey", tableName, strings.Join(schemaForeignKey.Columns, "_"))
}

func GenerateSqliteFKName(tableName string, schemaForeignKey *schemasv1alpha4.SqliteTableForeignKey) string {
	if schemaForeignKey.Name != "" {
		return schemaForeignKey.Name
	}

	return fmt.Sprintf("%s_%s_fkey", tableName, strings.Join(schemaForeignKey.Columns, "_"))
}
