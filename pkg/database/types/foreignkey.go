package types

import (
	"fmt"
	"strings"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
)

type ForeignKey struct {
	ChildColumns  []string
	ParentTable   string
	ParentColumns []string
	Name          string
}

func (fk *ForeignKey) Equals(other *ForeignKey) bool {
	// TODO

	return false
}

func ForeignKeyToSchemaForeignKey(foreignKey *ForeignKey) *schemasv1alpha2.SQLTableForeignKey {
	schemaForeignKey := schemasv1alpha2.SQLTableForeignKey{
		Columns: foreignKey.ChildColumns,
		References: schemasv1alpha2.SQLTableForeignKeyReferences{
			Table:   foreignKey.ParentTable,
			Columns: foreignKey.ParentColumns,
		},
		Name: foreignKey.Name,
	}

	return &schemaForeignKey
}

func SchemaForeignKeyToForeignKey(schemaForeignKey *schemasv1alpha2.SQLTableForeignKey) *ForeignKey {
	foreignKey := ForeignKey{
		ChildColumns:  schemaForeignKey.Columns,
		ParentTable:   schemaForeignKey.References.Table,
		ParentColumns: schemaForeignKey.References.Columns,
	}

	return &foreignKey
}

func GenerateFKName(tableName string, schemaForeignKey *schemasv1alpha2.SQLTableForeignKey) string {
	if schemaForeignKey.Name != "" {
		return schemaForeignKey.Name
	}

	return fmt.Sprintf("%s_%s_fkey", tableName, strings.Join(schemaForeignKey.Columns, "_"))
}
