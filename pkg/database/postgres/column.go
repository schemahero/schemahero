package postgres

import (
	"fmt"

	"github.com/lib/pq"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha2.SQLTableColumn) (*types.Column, error) {
	column := &types.Column{}

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

func postgresColumnAsInsert(column *schemasv1alpha2.SQLTableColumn) (string, error) {
	// Note, we don't always quote the column type becuase of how pg handles these two statement very differently:

	// 1. create table "users" ("id" "bigint","login" "varchar(255)","name" "varchar(255)")
	// 2. create table "users" ("id" bigint,"login" varchar(255),"name" varchar(255))

	// if the column type is a known (safe) type, pass it unquoted, else pass whatever we received as quoted
	postgresColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("%s %s", pq.QuoteIdentifier(column.Name), postgresColumn.DataType)

	if postgresColumn.Constraints != nil && postgresColumn.Constraints.NotNull != nil {
		if *postgresColumn.Constraints.NotNull == true {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha2.SQLTableColumn) (string, error) {
	columnFields, err := postgresColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf(`alter table %s add column %s`, pq.QuoteIdentifier(tableName), columnFields)

	return statement, nil
}
