package postgres

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
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

		statement := fmt.Sprintf(`insert into %s (%s) values (%s) on conflict do nothing`, tableName, strings.Join(cols, ", "), strings.Join(vals, ", "))
		statements = append(statements, statement)
	}

	return statements, nil
}

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := columnAsInsert(desiredColumn)
		if err != nil {
			return nil, err
		}
		columns = append(columns, columnFields)
	}

	if len(tableSchema.PrimaryKey) > 0 {
		primaryKeyColumns := []string{}
		for _, primaryKeyColumn := range tableSchema.PrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, pgx.Identifier{primaryKeyColumn}.Sanitize())
		}

		columns = append(columns, fmt.Sprintf("primary key (%s)", strings.Join(primaryKeyColumns, ", ")))
	}

	if len(tableSchema.Indexes) > 0 {
		for _, index := range tableSchema.Indexes {
			if index.IsUnique {
				uniqueColumns := []string{}
				for _, indexColumn := range index.Columns {
					uniqueColumns = append(uniqueColumns, pgx.Identifier{indexColumn}.Sanitize())
				}
				columns = append(columns, fmt.Sprintf("constraint %q unique (%s)", types.GeneratePostgresqlIndexName(tableName, index), strings.Join(uniqueColumns, ", ")))
			}
		}
	}

	if tableSchema.ForeignKeys != nil {
		for _, foreignKey := range tableSchema.ForeignKeys {
			columns = append(columns, foreignKeyConstraintClause(tableName, foreignKey))
		}
	}

	queries := []string{
		fmt.Sprintf(`create table %s (%s)`, pgx.Identifier{tableName}.Sanitize(), strings.Join(columns, ", ")),
	}

	// Add any triggers that are defined
	for _, trigger := range tableSchema.Triggers {
		statement, err := triggerCreateStatement(trigger, tableName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create trigger statement")
		}

		queries = append(queries, statement)
	}
	return queries, nil
}
