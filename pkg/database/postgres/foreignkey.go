package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveForeignKeyStatement(tableName string, foreignKey *types.ForeignKey) string {
	return fmt.Sprintf("alter table %s drop constraint %s", tableName, pgx.Identifier{foreignKey.Name}.Sanitize())
}

func AddForeignKeyStatement(tableName string, schemaForeignKey *schemasv1alpha4.PostgresqlTableForeignKey) string {
	return fmt.Sprintf("alter table %s add %s", tableName, foreignKeyConstraintClause(tableName, schemaForeignKey))
}

func foreignKeyConstraintClause(tableName string, schemaForeignKey *schemasv1alpha4.PostgresqlTableForeignKey) string {
	onDelete := ""
	if schemaForeignKey.OnDelete != "" {
		onDelete = fmt.Sprintf(" on delete %s", schemaForeignKey.OnDelete)
	}

	return fmt.Sprintf("constraint %s foreign key (%s) references %s (%s)%s",
		types.GeneratePostgresqlFKName(tableName, schemaForeignKey),
		strings.Join(SanitizeArray(schemaForeignKey.Columns), ", "),
		pgx.Identifier{schemaForeignKey.References.Table}.Sanitize(),
		strings.Join(SanitizeArray(schemaForeignKey.References.Columns), ", "),
		onDelete)
}
