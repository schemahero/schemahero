package mysql

import (
	"fmt"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func columnsMatch(col1 *types.Column, col2 *types.Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	col1Constraints, col2Constraints := col1.Constraints, col2.Constraints
	if col1Constraints == nil {
		col1Constraints = &types.ColumnConstraints{}
	}
	if col2Constraints == nil {
		col2Constraints = &types.ColumnConstraints{}
	}

	// TODO: default

	return types.NotNullConstraintEquals(col1Constraints.NotNull, col2Constraints.NotNull)
}

func AlterColumnStatement(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha2.SQLTableColumn, existingColumn *types.Column) (string, error) {
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

			stmt := AlterModifyColumnStatement{
				TableName:  tableName,
				ColumnName: column.Name,
				DataType:   column.DataType,
				Default:    column.ColumnDefault,
			}

			isPrimaryKey := false
			for _, primaryKey := range primaryKeys {
				if column.Name == primaryKey {
					isPrimaryKey = true
				}
			}

			if column.Constraints != nil && column.Constraints.NotNull != nil && *column.Constraints.NotNull {
				stmt.NotNull = column.Constraints.NotNull
			} else if !isPrimaryKey && existingColumn.Constraints != nil && existingColumn.Constraints.NotNull != nil && *existingColumn.Constraints.NotNull {
				stmt.NotNull = new(bool)
			}

			return stmt.String(), nil
		}
	}

	// wasn't found as a desired column, so drop
	return fmt.Sprintf("alter table `%s` drop column `%s`", tableName, existingColumn.Name), nil
}
