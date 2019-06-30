package postgres

import (
	"fmt"
	"strings"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveIndexStatement(tableName string, index *types.Index) string {
	return fmt.Sprintf("alter table %s drop index %s", tableName, index.Name)
}

func AddIndexStatement(tableName string, schemaIndex *schemasv1alpha2.SQLTableIndex) string {
	unique := ""
	if schemaIndex.IsUnique {
		unique = "unique "
	}

	name := schemaIndex.Name
	if name == "" {
		name = types.GenerateIndexName(tableName, schemaIndex)
	}

	return fmt.Sprintf("create %sindex %s on %s (%s)",
		unique,
		name,
		tableName,
		strings.Join(schemaIndex.Columns, ", "))
}

func RenameIndexStatement(tableName string, index *types.Index, schemaIndex *schemasv1alpha2.SQLTableIndex) string {
	return fmt.Sprintf("alter index %s rename to %s", index.Name, schemaIndex.Name)
}
