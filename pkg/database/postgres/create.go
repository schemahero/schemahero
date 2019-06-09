package postgres

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func CreateTableStatement(tableName string, tableSchema *schemasv1alpha2.SQLTableSchema) (string, error) {
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

	if tableSchema.ForeignKeys != nil {
		for _, foreignKey := range tableSchema.ForeignKeys {
			columns = append(columns, fmt.Sprintf("constraint %q foreign key (%s) references %s (%s)",
				types.GenerateFKName(tableName, foreignKey),
				strings.Join(foreignKey.Columns, ", "),
				foreignKey.References.Table,
				strings.Join(foreignKey.References.Columns, ", ")))
		}
	}

	query := fmt.Sprintf(`create table %s (%s)`, pq.QuoteIdentifier(tableName), strings.Join(columns, ", "))

	return query, nil
}
