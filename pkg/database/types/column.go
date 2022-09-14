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
	if a == nil || !*a {
		return b == nil || !*b
	}
	return b != nil && *b
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

func ColumnToMysqlSchemaColumn(column *Column) (*schemasv1alpha4.MysqlTableColumn, error) {
	schemaColumn := &schemasv1alpha4.MysqlTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha4.MysqlTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.Attributes != nil {
		schemaColumn.Attributes = &schemasv1alpha4.MysqlTableColumnAttributes{
			AutoIncrement: column.Attributes.AutoIncrement,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	schemaColumn.Charset = column.Charset
	schemaColumn.Collation = column.Collation

	return schemaColumn, nil
}

func ColumnToPostgresqlSchemaColumn(column *Column) (*schemasv1alpha4.PostgresqlTableColumn, error) {
	schemaColumn := &schemasv1alpha4.PostgresqlTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha4.PostgresqlTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.Attributes != nil {
		schemaColumn.Attributes = &schemasv1alpha4.PostgresqlTableColumnAttributes{
			AutoIncrement: column.Attributes.AutoIncrement,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	return schemaColumn, nil
}

func ColumnToRqliteSchemaColumn(column *Column) (*schemasv1alpha4.RqliteTableColumn, error) {
	schemaColumn := &schemasv1alpha4.RqliteTableColumn{
		Name: column.Name,
		Type: column.DataType,
	}

	if column.Constraints != nil {
		schemaColumn.Constraints = &schemasv1alpha4.RqliteTableColumnConstraints{
			NotNull: column.Constraints.NotNull,
		}
	}

	if column.Attributes != nil {
		schemaColumn.Attributes = &schemasv1alpha4.RqliteTableColumnAttributes{
			AutoIncrement: column.Attributes.AutoIncrement,
		}
	}

	schemaColumn.Default = column.ColumnDefault

	return schemaColumn, nil
}
