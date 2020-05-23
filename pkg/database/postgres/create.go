package postgres

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func CreateTableStatement(tableName string, tableSchema *schemasv1alpha4.SQLTableSchema) (string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := postgresColumnAsInsert(desiredColumn)
		if err != nil {
			return "", err
		}
		columns = append(columns, columnFields)
	}

	if len(tableSchema.PrimaryKey) > 0 {
		primaryKeyColumns := []string{}
		for _, primaryKeyColumn := range tableSchema.PrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, pq.QuoteIdentifier(primaryKeyColumn))
		}

		columns = append(columns, fmt.Sprintf("primary key (%s)", strings.Join(primaryKeyColumns, ", ")))
	}

	if len(tableSchema.Indexes) > 0 {
		for _, index := range tableSchema.Indexes {
			if index.IsUnique {
				uniqueColumns := []string{}
				for _, indexColumn := range index.Columns {
					uniqueColumns = append(uniqueColumns, pq.QuoteIdentifier(indexColumn))
				}
				columns = append(columns, fmt.Sprintf("constraint %q unique (%s)", types.GenerateIndexName(tableName, index), strings.Join(uniqueColumns, ", ")))
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

	query := fmt.Sprintf(`create table %s (%s)`, pq.QuoteIdentifier(tableName), strings.Join(columns, ", "))

	return query, nil
}
