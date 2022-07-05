package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatements(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha4.PostgresqlTableColumn, existingColumn *types.Column) ([]string, error) {
	alterStatement := fmt.Sprintf("alter column %s", pgx.Identifier{existingColumn.Name}.Sanitize())

	// this could be an alter or a drop column command
	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := schemaColumnToColumn(desiredColumn)
			if err != nil {
				return nil, err
			}

			if columnsMatch(*existingColumn, *column) {
				return []string{}, nil
			}

			// If the request is to modify a column to add a not null contraint to an existing column
			// handle that part here
			if column.Constraints != nil && column.Constraints.NotNull != nil && *column.Constraints.NotNull {
				isAddingNotNull := false
				if existingColumn.Constraints == nil {
					isAddingNotNull = true
				} else if existingColumn.Constraints.NotNull == nil {
					isAddingNotNull = true
				} else if !*existingColumn.Constraints.NotNull {
					isAddingNotNull = true
				}

				if isAddingNotNull {
					// the best plan here is:
					//   1. add default
					//   2. update values with default
					//   3. set not null

					statements := []string{}

					// add default
					if column.ColumnDefault != nil {
						if existingColumn.ColumnDefault == nil || *existingColumn.ColumnDefault != *column.ColumnDefault {
							localStatement := fmt.Sprintf("alter table %s alter column %s set default '%s'",
								pgx.Identifier{tableName}.Sanitize(),
								pgx.Identifier{existingColumn.Name}.Sanitize(),
								*column.ColumnDefault)
							statements = append(statements, localStatement)
						}
					}

					// update existing values
					if column.ColumnDefault != nil {
						localStatement := fmt.Sprintf("update %s set %s='%s' where %s is null",
							pgx.Identifier{tableName}.Sanitize(),
							pgx.Identifier{existingColumn.Name}.Sanitize(),
							*column.ColumnDefault,
							pgx.Identifier{existingColumn.Name}.Sanitize())
						statements = append(statements, localStatement)
					}

					// set not null
					localStatement := fmt.Sprintf("alter table %s alter column %s set not null",
						pgx.Identifier{tableName}.Sanitize(),
						pgx.Identifier{existingColumn.Name}.Sanitize())
					statements = append(statements, localStatement)

					return statements, nil
				}
			}

			changes := []string{}
			if column.DataType != "serial" && column.DataType != "bigserial" {
				if existingColumn.DataType != column.DataType {
					changes = append(changes, fmt.Sprintf("%s type %s", alterStatement, column.DataType))
				} else if column.DataType == existingColumn.DataType {
					if column.IsArray != existingColumn.IsArray {
						changes = append(changes, fmt.Sprintf("%s type %s[] using %s::%s[]", alterStatement, column.DataType, pgx.Identifier{existingColumn.Name}.Sanitize(), column.DataType))
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
					isPrimaryKey := false
					for _, primaryKey := range primaryKeys {
						if column.Name == primaryKey {
							isPrimaryKey = true
						}
					}

					if !isPrimaryKey {
						if existingColumn.Constraints != nil && existingColumn.Constraints.NotNull != nil && *existingColumn.Constraints.NotNull {
							if column.Constraints == nil || column.Constraints.NotNull == nil || !*column.Constraints.NotNull {
								changes = append(changes, fmt.Sprintf("%s drop not null", alterStatement))
							}
						}
					}
				}
			}

			if len(changes) == 0 {
				// no changes
				return []string{}, nil
			}

			return []string{fmt.Sprintf(`alter table %s %s`, pgx.Identifier{tableName}.Sanitize(), strings.Join(changes, ", "))}, nil
		}
	}

	return []string{fmt.Sprintf(`alter table %s drop column %s`, pgx.Identifier{tableName}.Sanitize(), pgx.Identifier{existingColumn.Name}.Sanitize())}, nil
}

func columnsMatch(col1 types.Column, col2 types.Column) bool {
	// for time and timestamp comparisons, we know that postgres
	// defaults to without time zone, so let's normalize a case
	// where someone is relying on the default and specifying it
	if col1.DataType == "timestamp" {
		col1.DataType = "timestamp without time zone"
	}
	if col2.DataType == "timestamp" {
		col2.DataType = "timestamp without time zone"
	}

	if col1.DataType == "time" {
		col1.DataType = "time without time zone"
	}
	if col2.DataType == "time" {
		col2.DataType = "time without time zone"
	}

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

	return types.BoolsEqual(col1Constraints.NotNull, col2Constraints.NotNull)
}
