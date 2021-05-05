package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

// schemaColumnToColumn converts the requested type from the v1alpha4 schema to a postgres schema type
func schemaColumnToColumn(schemaColumn *schemasv1alpha4.PostgresqlTableColumn) (*types.Column, error) {
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

	// split on the "[" character, which is only present in arrays
	requestedType = strings.Split(requestedType, "[")[0]
	// if the first element after splitting is not the entire string, it's an array
	column.IsArray = len(requestedType) < len(schemaColumn.Type)

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

func columnAsInsert(column *schemasv1alpha4.PostgresqlTableColumn) (string, error) {
	// Note, we don't always quote the column type because of how pg handles these two statement very differently:

	// 1. create table "users" ("id" "bigint","login" "varchar(255)","name" "varchar(255)")
	// 2. create table "users" ("id" bigint,"login" varchar(255),"name" varchar(255))

	// if the column type is a known (safe) type, pass it unquoted, else pass whatever we received as quoted
	postgresColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	arraySpecifier := ""
	if postgresColumn.IsArray {
		arraySpecifier = "[]"
	}

	formatted := fmt.Sprintf("%s %s%s", pgx.Identifier{column.Name}.Sanitize(), postgresColumn.DataType, arraySpecifier)

	if postgresColumn.Constraints != nil && postgresColumn.Constraints.NotNull != nil {
		if *postgresColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if postgresColumn.ColumnDefault != nil {
		value := stripOIDClass(*postgresColumn.ColumnDefault)
		formatted = fmt.Sprintf("%s default '%s'", formatted, value)
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.PostgresqlTableColumn) (string, error) {
	columnFields, err := columnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf(`alter table %s add column %s`, pgx.Identifier{tableName}.Sanitize(), columnFields)

	return statement, nil
}
