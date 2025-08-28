package sqlite

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.SqliteTableColumn) (*types.Column, error) {
	column := &types.Column{
		Name:          schemaColumn.Name,
		DataType:      schemaColumn.Type,
		ColumnDefault: schemaColumn.Default,
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

func sqliteColumnAsInsert(column *schemasv1alpha4.SqliteTableColumn) (string, error) {
	sqliteColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf(`"%s" %s`, column.Name, sqliteColumn.DataType)

	if sqliteColumn.Collation != "" {
		formatted = fmt.Sprintf("%s collate %s", formatted, sqliteColumn.Collation)
	}

	if sqliteColumn.Constraints != nil && sqliteColumn.Constraints.NotNull != nil {
		if *sqliteColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if sqliteColumn.Attributes != nil && sqliteColumn.Attributes.AutoIncrement != nil && *sqliteColumn.Attributes.AutoIncrement {
		formatted = fmt.Sprintf("%s autoincrement", formatted)
	}

	if sqliteColumn.ColumnDefault != nil {
		formatted = fmt.Sprintf("%s default '%s'", formatted, *sqliteColumn.ColumnDefault)
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.SqliteTableColumn) (string, error) {
	columnFields, err := sqliteColumnAsInsert(desiredColumn)
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
