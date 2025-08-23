package rqlite

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveIndexStatement(tableName string, index *types.Index) string {
	if index.IsUnique {
		return fmt.Sprintf("drop index if exists %s", index.Name)
	}
	return fmt.Sprintf("drop index %s", index.Name)
}

func AddIndexStatement(tableName string, schemaIndex *schemasv1alpha4.RqliteTableIndex) string {
	unique := ""
	if schemaIndex.IsUnique {
		unique = "unique "
	}

	name := schemaIndex.Name
	if name == "" {
		name = types.GenerateRqliteIndexName(tableName, schemaIndex)
	}

	return fmt.Sprintf("create %sindex %s on %s (%s)",
		unique,
		name,
		tableName,
		strings.Join(schemaIndex.Columns, ", "))
}
