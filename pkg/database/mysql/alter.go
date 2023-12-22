package mysql

import (
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func AlterColumnStatements(tableName string, primaryKeys []string, desiredColumns []*schemasv1alpha4.MysqlTableColumn, existingColumn *types.Column, defaultCharset string, defaultCollation string) ([]string, error) {
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

			if columnsMatch(*existingColumn, *column, defaultCharset, defaultCollation) {
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

func columnsMatch(existingCol types.Column, specCol types.Column, defaultCharset string, defaultCollation string) bool {
	if existingCol.DataType != specCol.DataType {
		return false
	}

	if existingCol.Charset == "" {
		existingCol.Charset = defaultCharset
	}
	if specCol.Charset == "" {
		specCol.Charset = defaultCharset
	}

	// Don't override collation in spec if it's not set. MySQL will select collation based on charset.
	if existingCol.Collation == "" {
		existingCol.Collation = defaultCollation
	}

	// now that we've applied defaults, let's see if they actually are different
	if existingCol.Charset != specCol.Charset {
		return false
	}

	if specCol.Collation != "" && existingCol.Collation != specCol.Collation {
		return false
	}

	if existingCol.ColumnDefault != nil && specCol.ColumnDefault == nil {
		return false
	} else if existingCol.ColumnDefault == nil && specCol.ColumnDefault != nil {
		return false
	} else if existingCol.ColumnDefault != nil && specCol.ColumnDefault != nil && *existingCol.ColumnDefault != *specCol.ColumnDefault {
		return false
	}

	col1Constraints, col2Constraints := existingCol.Constraints, specCol.Constraints
	if col1Constraints == nil {
		col1Constraints = &types.ColumnConstraints{}
	}
	if col2Constraints == nil {
		col2Constraints = &types.ColumnConstraints{}
	}

	if !types.BoolsEqual(col1Constraints.NotNull, col2Constraints.NotNull) {
		return false
	}

	col1Attributes, col2Attributes := existingCol.Attributes, specCol.Attributes
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
