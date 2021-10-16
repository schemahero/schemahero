package mysql

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := mysqlColumnAsInsert(desiredColumn)
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

	for _, index := range tableSchema.Indexes {
		columns = append(columns, indexClause(tableName, index))
	}

	query := fmt.Sprintf("create table `%s` (%s)", tableName, strings.Join(columns, ", "))

	if tableSchema.DefaultCharset != "" {
		query = fmt.Sprintf("%s default character set %s", query, tableSchema.DefaultCharset)
	}
	if tableSchema.Collation != "" {
		query = fmt.Sprintf("%s collate %s", query, tableSchema.Collation)
	}

	return []string{query}, nil
}
