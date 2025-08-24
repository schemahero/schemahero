package cassandra

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.CassandraColumn) (*types.Column, error) {
	// TODO validate types

	column := &types.Column{
		Name: schemaColumn.Name,
	}

	if schemaColumn.IsStatic != nil {
		column.IsStatic = *schemaColumn.IsStatic
	}

	unaliasedColumnType := unaliasUnparameterizedColumnType(schemaColumn.Type)
	if unaliasedColumnType != "" {
		column.DataType = unaliasedColumnType
	}

	if column.DataType == "" {
		column.DataType = schemaColumn.Type
	}

	return column, nil
}

func cassandraColumnAsInsert(column *schemasv1alpha4.CassandraColumn) (string, error) {
	// TODO before merge!  find the right gocql sanitize methods to call

	result := fmt.Sprintf("%s %s", column.Name, column.Type)
	if column.IsStatic != nil && *column.IsStatic {
		result = fmt.Sprintf("%s static", result)
	}

	return result, nil
}

func InsertColumnStatement(keyspace string, tableName string, desiredColumn *schemasv1alpha4.CassandraColumn) (string, error) {
	columnFields, err := cassandraColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf(`alter table "%s.%s" add %s`, keyspace, tableName, columnFields)

	return statement, nil
}
