package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := postgresColumnAsInsert(desiredColumn)
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
			} else {
				// non unique indexes are not supported in fixtures
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

	return queries, nil
}
