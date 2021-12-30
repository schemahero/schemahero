package mysql

import (
	"fmt"
	"strconv"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func SeedDataStatements(tableName string, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	statements := []string{}

	for _, row := range seedData.Rows {
		cols := []string{}
		vals := []string{}
		updateVals := []string{}
		for _, col := range row.Columns {
			cols = append(cols, col.Column)
			if col.Value.Int != nil {
				vals = append(vals, strconv.Itoa(*col.Value.Int))
				updateVals = append(updateVals, fmt.Sprintf("%s=%s", col.Column, strconv.Itoa(*col.Value.Int)))
			} else if col.Value.Str != nil {
				// handle multiline strings
				if strings.Contains(*col.Value.Str, "\n") {
					builder := []string{
						"CONCAT_WS(CHAR(10 using utf8)",
					}
					for _, s := range strings.Split(*col.Value.Str, "\n") {
						builder = append(builder, fmt.Sprintf("'%s'", s))
					}
					vals = append(vals, fmt.Sprintf("%s)", strings.Join(builder, ", ")))
					updateVals = append(updateVals, fmt.Sprintf("%s=%s", col.Column, fmt.Sprintf("%s)", strings.Join(builder, ", "))))
				} else {
					vals = append(vals, fmt.Sprintf("'%s'", *col.Value.Str))
					updateVals = append(updateVals, fmt.Sprintf("%s=%s", col.Column, fmt.Sprintf("'%s'", *col.Value.Str)))
				}
			}
		}

		statement := fmt.Sprintf(`insert into %s (%s) values (%s) on duplicate key update %s`, tableName, strings.Join(cols, ", "), strings.Join(vals, ", "), strings.Join(updateVals, ", "))
		statements = append(statements, statement)
	}

	return statements, nil
}

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
