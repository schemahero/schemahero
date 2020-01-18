package types

import (
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
)

type ColumnConstraints struct {
	NotNull *bool
}

func NotNullConstraintEquals(a, b *bool) bool {
	if a == nil || bool(*a) == false {
		return b == nil || bool(*b) == false
	}
	return b != nil && bool(*b) == true
}

type Column struct {
	Name          string
	DataType      string
	ColumnDefault *string
	Constraints   *ColumnConstraints
	IsArray       bool
}

func ColumnToSchemaColumn(column *Column) (*schemasv1alpha3.SQLTableColumn, error) {
	schemaColumn := &schemasv1alpha3.SQLTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha3.SQLTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	return schemaColumn, nil
}
