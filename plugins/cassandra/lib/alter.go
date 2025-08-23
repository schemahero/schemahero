package cassandra

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatements(keyspace string, tableName string, desiredColumns []*schemasv1alpha4.CassandraColumn, existingColumn *types.Column) ([]string, error) {
	alterStatement := fmt.Sprintf("alter column %s", existingColumn.Name)

	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := schemaColumnToColumn(desiredColumn)
			if err != nil {
				return nil, err
			}

			if columnsMatch(*existingColumn, *column) {
				return []string{}, nil
			}

			changes := []string{}
			if existingColumn.DataType != column.DataType {
				changes = append(changes, fmt.Sprintf("%s type %s", alterStatement, column.DataType))
			}

			if len(changes) == 0 {
				return []string{}, nil
			}

			return []string{fmt.Sprintf(`alter table "%s.%s" %s`, keyspace, tableName, strings.Join(changes, ", "))}, nil
		}
	}

	return []string{fmt.Sprintf(`alter table "%s.%s" drop column %s`, keyspace, tableName, existingColumn.Name)}, nil
}

func columnsMatch(col1 types.Column, col2 types.Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	if col1.IsStatic != col2.IsStatic {
		return false
	}

	return true
}
