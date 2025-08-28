package mysql

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.MysqlTableColumn) (*types.Column, error) {
	column := &types.Column{
		Name:          schemaColumn.Name,
		ColumnDefault: schemaColumn.Default,
		Charset:       schemaColumn.Charset,
		Collation:     schemaColumn.Collation,
	}

	if schemaColumn.Constraints != nil {
		column.Constraints = &types.ColumnConstraints{
			NotNull: schemaColumn.Constraints.NotNull,
		}
	}

	if schemaColumn.Attributes != nil {
		column.Attributes = &types.ColumnAttributes{
			AutoIncrement: schemaColumn.Attributes.AutoIncrement,
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

func mysqlColumnAsInsert(column *schemasv1alpha4.MysqlTableColumn) (string, error) {
	mysqlColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("`%s` %s", column.Name, mysqlColumn.DataType)

	if mysqlColumn.Charset != "" {
		formatted = fmt.Sprintf("%s character set %s", formatted, mysqlColumn.Charset)
	}
	if mysqlColumn.Collation != "" {
		formatted = fmt.Sprintf("%s collate %s", formatted, mysqlColumn.Collation)
	}

	if mysqlColumn.Constraints != nil && mysqlColumn.Constraints.NotNull != nil {
		if *mysqlColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if mysqlColumn.Attributes != nil && mysqlColumn.Attributes.AutoIncrement != nil && *mysqlColumn.Attributes.AutoIncrement {
		formatted = fmt.Sprintf("%s auto_increment", formatted)
	}

	if mysqlColumn.ColumnDefault != nil {
		quoteDefaultValue := true

		if mysqlColumn.DataType == "datetime" || mysqlColumn.DataType == "timestamp" {
			if *mysqlColumn.ColumnDefault == "CURRENT_TIMESTAMP" {
				quoteDefaultValue = false
			}
		}

		if quoteDefaultValue {
			formatted = fmt.Sprintf("%s default '%s'", formatted, *mysqlColumn.ColumnDefault)
		} else {
			formatted = fmt.Sprintf("%s default %s", formatted, *mysqlColumn.ColumnDefault)
		}
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.MysqlTableColumn) (string, error) {
	columnFields, err := mysqlColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf("alter table `%s` add column %s", tableName, columnFields)

	return statement, nil
}
