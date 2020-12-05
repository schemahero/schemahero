package sqlite

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.SqliteTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := sqliteColumnAsInsert(desiredColumn)
		if err != nil {
			return nil, err
		}
		columns = append(columns, columnFields)
	}

	if len(tableSchema.PrimaryKey) > 0 {
		primaryKeyColumns := []string{}
		for _, primaryKeyColumn := range tableSchema.PrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, fmt.Sprintf("`%s`", primaryKeyColumn))
		}

		columns = append(columns, fmt.Sprintf("primary key (%s)", strings.Join(primaryKeyColumns, ", ")))
	}

	if tableSchema.ForeignKeys != nil {
		for _, foreignKey := range tableSchema.ForeignKeys {
			columns = append(columns, foreignKeyConstraintClause(tableName, foreignKey))
		}

	}

	query := fmt.Sprintf("create table `%s` (%s)", tableName, strings.Join(columns, ", "))

	return []string{query}, nil
}
