package mysql

import (
	"fmt"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha2.SQLTableColumn) (*types.Column, error) {
	column := &types.Column{
		Name:          schemaColumn.Name,
		ColumnDefault: schemaColumn.Default,
	}

	if schemaColumn.Constraints != nil {
		column.Constraints = &types.ColumnConstraints{
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

func mysqlColumnAsInsert(column *schemasv1alpha2.SQLTableColumn) (string, error) {
	mysqlColumn, err := schemaColumnToColumn(column)
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

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha2.SQLTableColumn) (string, error) {
	columnFields, err := mysqlColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf("alter table `%s` add column %s", tableName, columnFields)

	return statement, nil
}
