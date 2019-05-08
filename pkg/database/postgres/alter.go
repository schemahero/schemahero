package postgres

import (
	"fmt"

	"github.com/lib/pq"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
)

func columnsMatch(col1 *Column, col2 *Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	if col1.Constraints == nil && col2.Constraints != nil {
		return false
	}

	if col1.Constraints != nil && col2.Constraints == nil {
		return false
	}

	if col1.Constraints != nil && col2.Constraints != nil {
		if col1.Constraints.NotNull != col2.Constraints.NotNull {
			return false
		}
	}

	return true
}

func AlterColumnStatement(tableName string, desiredColumns []*schemasv1alpha1.PostgresTableColumn, existingColumn *Column) (string, error) {
	// this could be an alter or a drop column command
	columnStatement := ""
	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := schemaColumnToPostgresColumn(desiredColumn)
			if err != nil {
				return "", err
			}

			if columnsMatch(existingColumn, column) {
				return "", nil
			}

			columnStatement = fmt.Sprintf(`alter table %s alter column %s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name))

			if existingColumn.DataType != column.DataType {
				columnStatement = fmt.Sprintf("%s type %s", columnStatement, column.DataType)
			}

			if column.Constraints != nil {
				if existingColumn.Constraints != nil {
					if column.Constraints.NotNull && !existingColumn.Constraints.NotNull {
						columnStatement = fmt.Sprintf("%s set not null", columnStatement)
					} else if !column.Constraints.NotNull && existingColumn.Constraints.NotNull {
						columnStatement = fmt.Sprintf("%s drop not null", columnStatement)
					}
				}
			}
		}
	}
	if columnStatement == "" {
		// wasn't found as a desired column, so drop
		columnStatement = fmt.Sprintf(`alter table %s drop column %s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name))
	}

	return columnStatement, nil
}
