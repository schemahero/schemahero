package mysql

import (
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatements(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha4.MysqlSQLTableColumn, existingColumn *types.Column) ([]string, error) {
	// this could be an alter or a drop column command
	for _, desiredColumn := range desiredColumns {
		if desiredColumn.Name == existingColumn.Name {
			column, err := schemaColumnToColumn(desiredColumn)
			if err != nil {
				return nil, err
			}

			isPrimaryKey := false
			for _, primaryKey := range primaryKeys {
				if column.Name == primaryKey {
					isPrimaryKey = true
				}
			}

			// primary keys are always not null
			if isPrimaryKey {
				ensureColumnConstraintsNotNullTrue(column)
			}

			if columnsMatch(existingColumn, column) {
				return []string{}, nil
			}

			return AlterModifyColumnStatement{
				TableName:      tableName,
				ExistingColumn: *existingColumn,
				Column:         *column,
			}.DDL(), nil
		}
	}

	// wasn't found as a desired column, so drop
	return AlterDropColumnStatement{
		TableName: tableName,
		Column:    types.Column{Name: existingColumn.Name},
	}.DDL(), nil
}

func columnsMatch(col1 *types.Column, col2 *types.Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	if col1.Charset != col2.Charset {
		return false
	}

	if col1.Collation != col2.Collation {
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

	if !types.BoolsEqual(col1Constraints.NotNull, col2Constraints.NotNull) {
		return false
	}

	col1Attributes, col2Attributes := col1.Attributes, col2.Attributes
	if col1Attributes == nil {
		col1Attributes = &types.ColumnAttributes{}
	}
	if col2Attributes == nil {
		col2Attributes = &types.ColumnAttributes{}
	}

	if !types.BoolsEqual(col1Attributes.AutoIncrement, col2Attributes.AutoIncrement) {
		return false
	}

	return true
}

func ensureColumnConstraintsNotNullTrue(column *types.Column) {
	if column.Constraints == nil {
		column.Constraints = &types.ColumnConstraints{}
	}
	column.Constraints.NotNull = &trueValue
}
