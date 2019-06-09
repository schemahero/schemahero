package types

import (
	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
)

type ColumnConstraints struct {
	NotNull *bool
}

type Column struct {
	Name          string
	DataType      string
	ColumnDefault *string
	Constraints   *ColumnConstraints
}

func ColumnToSchemaColumn(column *Column) (*schemasv1alpha2.SQLTableColumn, error) {
	schemaColumn := &schemasv1alpha2.SQLTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha2.SQLTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.ColumnDefault != nil {
		schemaColumn.Default = *column.ColumnDefault
	}

	return schemaColumn, nil
}
