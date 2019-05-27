package mysql

import (
	"fmt"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
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

func MysqlColumnToSchemaColumn(column *Column) (*schemasv1alpha1.SQLTableColumn, error) {
	constraints := &schemasv1alpha1.SQLTableColumnConstraints{
		NotNull: column.Constraints.NotNull,
	}

	schemaColumn := &schemasv1alpha1.SQLTableColumn{
		Name:        column.Name,
		Type:        column.DataType,
		Constraints: constraints,
	}

	if column.ColumnDefault != nil {
		schemaColumn.Default = *column.ColumnDefault
	}

	return schemaColumn, nil
}

func schemaColumnToMysqlColumn(schemaColumn *schemasv1alpha1.SQLTableColumn) (*Column, error) {
	column := &Column{}

	if schemaColumn.Constraints != nil {
		column.Constraints = &ColumnConstraints{
			NotNull: schemaColumn.Constraints.NotNull,
		}
	}

	requestedType := schemaColumn.Type
	unaliasedColumnType := unaliasUnparameterizedColumnType(requestedType)
	if unaliasedColumnType != "" {
		requestedType = unaliasedColumnType
	}

	unaliasedColumnType = unaliasParameterizedColumnType(requestedType)
	if unaliasedColumnType != "" {
		requestedType = unaliasedColumnType
	}

	if !isParameterizedColumnType(requestedType) {
		column.DataType = requestedType
		return column, nil
	}

	columnType, err := maybeParseParameterizedColumnType(requestedType)
	if err != nil {
		return nil, err
	}

	if columnType != "" {
		column.DataType = columnType
		return column, nil
	}

	return nil, fmt.Errorf("unknown column type. cannot validate column type %q", schemaColumn.Type)
}

func mysqlColumnAsInsert(column *schemasv1alpha1.SQLTableColumn) (string, error) {
	mysqlColumn, err := schemaColumnToMysqlColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("`%s` %s", column.Name, mysqlColumn.DataType)

	if mysqlColumn.Constraints != nil && mysqlColumn.Constraints.NotNull != nil {
		if *mysqlColumn.Constraints.NotNull == true {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha1.SQLTableColumn) (string, error) {
	columnFields, err := mysqlColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf("alter table `%s` add column %s", tableName, columnFields)

	return statement, nil
}
