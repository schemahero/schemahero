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

func ForeignKeyToSchemaForeignKey(foreignKey *ForeignKey) *schemasv1alpha4.SQLTableForeignKey {
	schemaForeignKey := schemasv1alpha4.SQLTableForeignKey{
		Columns: foreignKey.ChildColumns,
		References: schemasv1alpha4.SQLTableForeignKeyReferences{
			Table:   foreignKey.ParentTable,
			Columns: foreignKey.ParentColumns,
		},
		Name:     foreignKey.Name,
		OnDelete: foreignKey.OnDelete,
	}

	return &schemaForeignKey
}

func SchemaForeignKeyToForeignKey(schemaForeignKey *schemasv1alpha4.SQLTableForeignKey) *ForeignKey {
	foreignKey := ForeignKey{
		ChildColumns:  schemaForeignKey.Columns,
		ParentTable:   schemaForeignKey.References.Table,
		ParentColumns: schemaForeignKey.References.Columns,
		Name:          schemaForeignKey.Name,
		OnDelete:      schemaForeignKey.OnDelete,
	}

	return &foreignKey
}

func GenerateFKName(tableName string, schemaForeignKey *schemasv1alpha4.SQLTableForeignKey) string {
	if schemaForeignKey.Name != "" {
		return schemaForeignKey.Name
	}

	return fmt.Sprintf("%s_%s_fkey", tableName, strings.Join(schemaForeignKey.Columns, "_"))
}
