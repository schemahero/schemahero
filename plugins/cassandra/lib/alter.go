package cassandra

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatements(keyspace string, tableName string, desiredColumns []*schemasv1alpha4.CassandraColumn, existingColumn *types.Column) ([]string, error) {
	
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
				// Cassandra syntax: ALTER TABLE ... ALTER column_name TYPE new_type
				changes = append(changes, fmt.Sprintf("alter %s type %s", existingColumn.Name, column.DataType))
			}

			if len(changes) == 0 {
				return []string{}, nil
			}

			// Don't include keyspace in table name since it's already set in the session
			_ = keyspace
			return []string{fmt.Sprintf(`alter table "%s" %s`, tableName, strings.Join(changes, ", "))}, nil
		}
	}

	// Don't include keyspace in table name since it's already set in the session
	_ = keyspace
	return []string{fmt.Sprintf(`alter table "%s" drop column %s`, tableName, existingColumn.Name)}, nil
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
