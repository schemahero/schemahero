package types

import (
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type ColumnConstraints struct {
	NotNull *bool
}

type ColumnAttributes struct {
	AutoIncrement *bool
}

func BoolsEqual(a, b *bool) bool {
	if a == nil || *a == false {
		return b == nil || *b == false
	}
	return b != nil && *b == true
}

type Column struct {
	Name          string
	DataType      string
	ColumnDefault *string
	Constraints   *ColumnConstraints
	Attributes    *ColumnAttributes
	IsArray       bool
	Charset       string
	Collation     string
	IsStatic      bool
}

func ColumnToMysqlSchemaColumn(column *Column) (*schemasv1alpha4.MysqlSQLTableColumn, error) {
	schemaColumn := &schemasv1alpha4.MysqlSQLTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha4.SQLTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.Attributes != nil {
		schemaColumn.Attributes = &schemasv1alpha4.SQLTableColumnAttributes{
			AutoIncrement: column.Attributes.AutoIncrement,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	schemaColumn.Charset = column.Charset
	schemaColumn.Collation = column.Collation

	return schemaColumn, nil
}

func ColumnToSchemaColumn(column *Column) (*schemasv1alpha4.SQLTableColumn, error) {
	schemaColumn := &schemasv1alpha4.SQLTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha4.SQLTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.Attributes != nil {
		schemaColumn.Attributes = &schemasv1alpha4.SQLTableColumnAttributes{
			AutoIncrement: column.Attributes.AutoIncrement,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	return schemaColumn, nil
}
