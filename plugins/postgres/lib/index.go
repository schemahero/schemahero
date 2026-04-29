package postgres

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveConstraintStatement(tableName string, index *types.Index) string {
	return fmt.Sprintf("alter table %s drop constraint %s", pgx.Identifier{tableName}.Sanitize(), pgx.Identifier{index.Name}.Sanitize())
}

func RemoveIndexStatement(tableName string, index *types.Index) string {
	if index.IsUnique {
		return fmt.Sprintf("drop index if exists %s", pgx.Identifier{index.Name}.Sanitize())
	}
	return fmt.Sprintf("drop index %s", pgx.Identifier{index.Name}.Sanitize())
}

func AddIndexStatement(tableName string, schemaIndex *schemasv1alpha4.PostgresqlTableIndex) string {
	unique := ""
	if schemaIndex.IsUnique {
		unique = "unique "
	}

	name := schemaIndex.Name
	if name == "" {
		name = types.GeneratePostgresqlIndexName(tableName, schemaIndex)
	}

	statement := fmt.Sprintf("create %sindex %s on %s (%s)",
		unique,
		name,
		tableName,
		strings.Join(schemaIndex.Columns, ", "))

	if schemaIndex.With != nil && len(schemaIndex.With) > 0 {
		keys := make([]string, 0, len(schemaIndex.With))
		for key := range schemaIndex.With {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		withClauses := make([]string, 0, len(schemaIndex.With))
		for _, key := range keys {
			withClauses = append(withClauses, fmt.Sprintf("%s = %s", key, schemaIndex.With[key]))
		}
		statement += fmt.Sprintf(" with (%s)", strings.Join(withClauses, ", "))
	}

	return statement
}

func RenameIndexStatement(tableName string, index *types.Index, schemaIndex *schemasv1alpha4.PostgresqlTableIndex) string {
	return fmt.Sprintf("alter index %s rename to %s", pgx.Identifier{index.Name}.Sanitize(), pgx.Identifier{schemaIndex.Name}.Sanitize())
}
