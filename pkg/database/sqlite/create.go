package sqlite

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
		for _, col := range row.Columns {
			cols = append(cols, col.Column)
			if col.Value.Int != nil {
				vals = append(vals, strconv.Itoa(*col.Value.Int))
			} else if col.Value.Str != nil {
				vals = append(vals, fmt.Sprintf("'%s'", *col.Value.Str))
			}
		}

		statement := fmt.Sprintf(`replace into %s (%s) values (%s)`, tableName, strings.Join(cols, ", "), strings.Join(vals, ", "))
		statements = append(statements, statement)
	}

	return statements, nil
}

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
			primaryKeyColumns = append(primaryKeyColumns, fmt.Sprintf(`"%s"`, primaryKeyColumn))
		}

		columns = append(columns, fmt.Sprintf("primary key (%s)", strings.Join(primaryKeyColumns, ", ")))
	}

	if tableSchema.ForeignKeys != nil {
		for _, foreignKey := range tableSchema.ForeignKeys {
			columns = append(columns, foreignKeyConstraintClause(tableName, foreignKey))
		}
	}

	query := fmt.Sprintf(`create table "%s" (%s)`, tableName, strings.Join(columns, ", "))
	if tableSchema.Strict {
		query = fmt.Sprintf("%s strict", query)
	}

	statements := []string{query}
	for _, index := range tableSchema.Indexes {
		statements = append(statements, AddIndexStatement(tableName, index))
	}

	return statements, nil
}
