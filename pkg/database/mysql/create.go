package mysql

import (
	"fmt"
	"strings"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
)

func CreateTableStatement(tableName string, tableSchema *schemasv1alpha1.SQLTableSchema) (string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := mysqlColumnAsInsert(desiredColumn)
		if err != nil {
			return "", err
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
			columns = append(columns, fmt.Sprintf("foreign key (%s) references %s (%s)", strings.Join(foreignKey.Columns, ", "), foreignKey.References.Table, strings.Join(foreignKey.References.Columns, ", ")))
		}

	}

	query := fmt.Sprintf("create table `%s` (%s)", tableName, strings.Join(columns, ", "))

	return query, nil
}
