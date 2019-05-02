package postgres

import (
	"fmt"

	"github.com/lib/pq"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
)

func AlterColumnStatement(tableName string, desiredColumns []*schemasv1alpha1.PostgresTableColumn, existingColumn *Column) (string, error) {
	// this could be an alter or a drop column command
	columnStatement := ""
	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := columnTypeToPostgresColumn(desiredColumn.Type)
			if err != nil {
				return "", err
			}

			if existingColumn.DataType == column.DataType {
				return "", nil
			}

			columnStatement = fmt.Sprintf(`alter table %s alter column %s type %s`,
				pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name), column.DataType)
		}
	}
	if columnStatement == "" {
		// wasn't found as a desired column, so drop
		columnStatement = fmt.Sprintf(`alter table %s drop column %s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name))
	}

	return columnStatement, nil
}
