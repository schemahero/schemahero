package sqlite

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.SqliteTableColumn) (*types.Column, error) {
	column := &types.Column{
		Name:          schemaColumn.Name,
		ColumnDefault: schemaColumn.Default,
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

func sqliteColumnAsInsert(column *schemasv1alpha4.SqliteTableColumn) (string, error) {
	sqliteColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("`%s` %s", column.Name, sqliteColumn.DataType)

	if sqliteColumn.Charset != "" {
		formatted = fmt.Sprintf("%s character set %s", formatted, sqliteColumn.Charset)
	}
	if sqliteColumn.Collation != "" {
		formatted = fmt.Sprintf("%s collate %s", formatted, sqliteColumn.Collation)
	}

	if sqliteColumn.Constraints != nil && sqliteColumn.Constraints.NotNull != nil {
		if *sqliteColumn.Constraints.NotNull == true {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if sqliteColumn.Attributes != nil && sqliteColumn.Attributes.AutoIncrement != nil && *sqliteColumn.Attributes.AutoIncrement {
		formatted = fmt.Sprintf("%s auto_increment", formatted)
	}

	if sqliteColumn.ColumnDefault != nil {
		quoteDefaultValue := true

		if sqliteColumn.DataType == "datetime" || sqliteColumn.DataType == "timestamp" {
			if *sqliteColumn.ColumnDefault == "CURRENT_TIMESTAMP" {
				quoteDefaultValue = false
			}
		}

		if quoteDefaultValue {
			formatted = fmt.Sprintf("%s default '%s'", formatted, *sqliteColumn.ColumnDefault)
		} else {
			formatted = fmt.Sprintf("%s default %s", formatted, *sqliteColumn.ColumnDefault)
		}
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.SqliteTableColumn) (string, error) {
	columnFields, err := sqliteColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf("alter table `%s` add column %s", tableName, columnFields)

	return statement, nil
}
