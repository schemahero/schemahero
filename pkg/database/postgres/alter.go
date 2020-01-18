package postgres

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatement(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha3.SQLTableColumn, existingColumn *types.Column) (string, error) {
	alterStatement := fmt.Sprintf("alter column %s", pq.QuoteIdentifier(existingColumn.Name))

	// this could be an alter or a drop column command
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
				changes = append(changes, fmt.Sprintf("%s type %s", alterStatement, column.DataType))
			} else if column.DataType == existingColumn.DataType {
				if column.IsArray != existingColumn.IsArray {
					changes = append(changes, fmt.Sprintf("%s type %s[] using %s::%s[]", alterStatement, column.DataType, pq.QuoteIdentifier(existingColumn.Name), column.DataType))
				}
			}

			if column.ColumnDefault != nil {
				if existingColumn.ColumnDefault == nil || *column.ColumnDefault != *existingColumn.ColumnDefault {
					changes = append(changes, fmt.Sprintf("%s set default '%s'", alterStatement, *column.ColumnDefault))
				}
			} else if existingColumn.ColumnDefault != nil {
				changes = append(changes, fmt.Sprintf("%s drop default", alterStatement))
			}

			// too much complexity below!
			if column.Constraints != nil || existingColumn.Constraints != nil {
				// Add not null
				if column.Constraints != nil && column.Constraints.NotNull != nil && *column.Constraints.NotNull == true {
					if existingColumn.Constraints != nil || existingColumn.Constraints.NotNull != nil {
						if *existingColumn.Constraints.NotNull == false {
							changes = append(changes, fmt.Sprintf("%s set not null", alterStatement))
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
							changes = append(changes, fmt.Sprintf("%s drop not null", alterStatement))
						}
					}
				}
			}

			if len(changes) == 0 {
				// no changes
				return "", nil
			}

			return fmt.Sprintf(`alter table %s %s`, pq.QuoteIdentifier(tableName), strings.Join(changes, ", ")), nil
		}
	}

	return fmt.Sprintf(`alter table %s drop column %s`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(existingColumn.Name)), nil
}

func columnsMatch(col1 *types.Column, col2 *types.Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	if col1.IsArray != col2.IsArray {
		return false
	}

	if col1.ColumnDefault != nil && col2.ColumnDefault == nil {
		return false
	} else if col1.ColumnDefault == nil && col2.ColumnDefault != nil {
		return false
	} else if col1.ColumnDefault != nil && col2.ColumnDefault != nil && *col1.ColumnDefault != *col2.ColumnDefault {
		return false
	}

	col1Constraints, col2Constraints := col1.Constraints, col2.Constraints
	if col1Constraints == nil {
		col1Constraints = &types.ColumnConstraints{}
	}
	if col2Constraints == nil {
		col2Constraints = &types.ColumnConstraints{}
	}

	return types.NotNullConstraintEquals(col1Constraints.NotNull, col2Constraints.NotNull)
}
