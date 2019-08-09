package postgres

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func columnsMatch(col1 *types.Column, col2 *types.Column) bool {
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

	// TODO: default

	return true
}

func AlterColumnStatement(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha2.SQLTableColumn, existingColumn *types.Column) (string, error) {
	// TODO: default

	// this could be an alter or a drop column command
	columnStatement := ""
	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := schemaColumnToColumn(desiredColumn)
			if err != nil {
				return "", err
			}

			if columnsMatch(existingColumn, column) {
				return "", nil
			}

			changes := []string{}
			if existingColumn.DataType != column.DataType {
				fmt.Printf("%s, %s\n", existingColumn.DataType, column.DataType)
				changes = append(changes, fmt.Sprintf("%s type %s", columnStatement, column.DataType))
			}

			// too much complexity below!
			if column.Constraints != nil || existingColumn.Constraints != nil {
				// Add not null
				if column.Constraints != nil && column.Constraints.NotNull != nil && *column.Constraints.NotNull == true {
					if existingColumn.Constraints != nil || existingColumn.Constraints.NotNull != nil {
						if *existingColumn.Constraints.NotNull == false {
							changes = append(changes, fmt.Sprintf("%s set not null", columnStatement))
						}
					}
				}

				isPrimaryKey := false
				for _, primaryKey := range primaryKeys {
					if column.Name == primaryKey {
						isPrimaryKey = true
					}
				}

				if !isPrimaryKey {
					if existingColumn.Constraints.NotNull != nil && *existingColumn.Constraints.NotNull == true {
						if column.Constraints == nil || column.Constraints.NotNull == nil || *column.Constraints.NotNull == false {
							changes = append(changes, fmt.Sprintf("%s drop not null", columnStatement))
						}
					}
				}
			}

			if len(changes) == 0 {
				// no changes
				return "", nil
			}

			columnStatement = fmt.Sprintf(`alter table %s alter column %s%s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name), strings.Join(changes, " "))

		}
	}
	if columnStatement == "" {
		// wasn't found as a desired column, so drop
		columnStatement = fmt.Sprintf(`alter table %s drop column %s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name))
	}

	return columnStatement, nil
}
