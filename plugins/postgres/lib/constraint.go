package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveConstrantStatement(tableName string, constraint *types.KeyConstraint) string {
	if constraint == nil {
		return ""
	}
	return fmt.Sprintf("alter table %s drop constraint %s", tableName, pgx.Identifier{constraint.Name}.Sanitize())
}

func AddConstrantStatement(tableName string, constraint *types.KeyConstraint) string {
	if constraint == nil {
		return ""
	}
	// `ALTER TABLE table_name ADD CONSTRAINT constraint_name PRIMARY KEY (index_col1, index_col2, ... index_col_n);
	return fmt.Sprintf(
		"alter table %s add constraint %s%s %s",
		tableName,
		constraint.GenerateName(tableName),
		primaryKeyClause(constraint),
		constraintColumnClause(constraint),
	)
}

func primaryKeyClause(constraint *types.KeyConstraint) string {
	if constraint.IsPrimary {
		return " primary key"
	}
	return ""
}

func constraintColumnClause(constraint *types.KeyConstraint) string {
	return fmt.Sprintf("(%s)", strings.Join(constraint.Columns, ", "))
}
