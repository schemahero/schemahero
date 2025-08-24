package rqlite

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func foreignKeyConstraintClause(tableName string, schemaForeignKey *schemasv1alpha4.RqliteTableForeignKey) string {
	onDelete := ""
	if schemaForeignKey.OnDelete != "" {
		onDelete = fmt.Sprintf(" on delete %s", schemaForeignKey.OnDelete)
	}

	return fmt.Sprintf("constraint %s foreign key (%s) references %s (%s)%s",
		types.GenerateRqliteFKName(tableName, schemaForeignKey),
		strings.Join(schemaForeignKey.Columns, ", "),
		schemaForeignKey.References.Table,
		strings.Join(schemaForeignKey.References.Columns, ", "),
		onDelete)
}
