package postgres

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveConstraintStatement(schema string, tableName string, index *types.Index) string {
	if schema != "" && schema != "public" {
		return fmt.Sprintf("alter table %s drop constraint %s", pgx.Identifier{schema, tableName}.Sanitize(), pgx.Identifier{index.Name}.Sanitize())
	}
	return fmt.Sprintf("alter table %s drop constraint %s", pgx.Identifier{tableName}.Sanitize(), pgx.Identifier{index.Name}.Sanitize())
}

func RemoveIndexStatement(schema string, tableName string, index *types.Index) string {
	if schema != "" && schema != "public" {
		if index.IsUnique {
			return fmt.Sprintf("drop index if exists %s", pgx.Identifier{schema, index.Name}.Sanitize())
		}
		return fmt.Sprintf("drop index %s", pgx.Identifier{schema, index.Name}.Sanitize())
	}
	if index.IsUnique {
		return fmt.Sprintf("drop index if exists %s", pgx.Identifier{index.Name}.Sanitize())
	}
	return fmt.Sprintf("drop index %s", pgx.Identifier{index.Name}.Sanitize())
}

func AddIndexStatement(schema string, tableName string, schemaIndex *schemasv1alpha4.PostgresqlTableIndex) string {
	unique := ""
	if schemaIndex.IsUnique {
		unique = "unique "
	}

	name := schemaIndex.Name
	if name == "" {
		name = types.GeneratePostgresqlIndexName(tableName, schemaIndex)
	}

	if schema != "" && schema != "public" {
		statement := fmt.Sprintf("create %sindex %s on %s (%s)",
			unique,
			pgx.Identifier{name}.Sanitize(),
			pgx.Identifier{schema, tableName}.Sanitize(),
			strings.Join(schemaIndex.Columns, ", "))

		if schemaIndex.With != nil && len(schemaIndex.With) > 0 {
			statement += buildWithClause(schemaIndex.With)
		}

		return statement
	}

	statement := fmt.Sprintf("create %sindex %s on %s (%s)",
		unique,
		name,
		tableName,
		strings.Join(schemaIndex.Columns, ", "))

	if schemaIndex.With != nil && len(schemaIndex.With) > 0 {
		statement += buildWithClause(schemaIndex.With)
	}

	return statement
}

func buildWithClause(with map[string]string) string {
	keys := make([]string, 0, len(with))
	for key := range with {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	withClauses := make([]string, 0, len(with))
	for _, key := range keys {
		withClauses = append(withClauses, fmt.Sprintf("%s = %s", key, with[key]))
	}
	return fmt.Sprintf(" with (%s)", strings.Join(withClauses, ", "))
}

func RenameIndexStatement(tableName string, index *types.Index, schemaIndex *schemasv1alpha4.PostgresqlTableIndex) string {
	return fmt.Sprintf("alter index %s rename to %s", pgx.Identifier{index.Name}.Sanitize(), pgx.Identifier{schemaIndex.Name}.Sanitize())
}
