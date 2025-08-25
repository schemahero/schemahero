package rqlite

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.RqliteTableColumn) (*types.Column, error) {
	// Handle the sentinel value for empty string defaults
	var columnDefault *string
	if schemaColumn.Default != nil {
		defaultValue := *schemaColumn.Default
		if defaultValue == "__SCHEMAHERO_EMPTY_STRING_DEFAULT__" {
			emptyStr := ""
			columnDefault = &emptyStr
		} else {
			columnDefault = schemaColumn.Default
		}
	}
	
	column := &types.Column{
		Name:          schemaColumn.Name,
		DataType:      schemaColumn.Type,
		ColumnDefault: columnDefault,
	}

	if schemaColumn.Constraints != nil {
		column.Constraints = &types.ColumnConstraints{
			NotNull: schemaColumn.Constraints.NotNull,
		}
	}

	if schemaColumn.Attributes != nil {
		column.Attributes = &types.ColumnAttributes{
			AutoIncrement: schemaColumn.Attributes.AutoIncrement,
		}
	}

	return column, nil
}

func rqliteColumnAsInsert(column *schemasv1alpha4.RqliteTableColumn) (string, error) {
	rqliteColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf(`"%s" %s`, column.Name, rqliteColumn.DataType)

	if rqliteColumn.Collation != "" {
		formatted = fmt.Sprintf("%s collate %s", formatted, rqliteColumn.Collation)
	}

	if rqliteColumn.Constraints != nil && rqliteColumn.Constraints.NotNull != nil {
		if *rqliteColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if rqliteColumn.Attributes != nil && rqliteColumn.Attributes.AutoIncrement != nil && *rqliteColumn.Attributes.AutoIncrement {
		formatted = fmt.Sprintf("%s autoincrement", formatted)
	}

	if rqliteColumn.ColumnDefault != nil {
		formatted = fmt.Sprintf("%s default '%s'", formatted, *rqliteColumn.ColumnDefault)
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.RqliteTableColumn) (string, error) {
	columnFields, err := rqliteColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf(`alter table "%s" add column %s`, tableName, columnFields)

	return statement, nil
}

func DropColumnStatement(tableName string, existingColumn types.Column) (string, error) {
	statement := fmt.Sprintf(`alter table "%s" drop column "%s"`, tableName, existingColumn.Name)
	return statement, nil
}
