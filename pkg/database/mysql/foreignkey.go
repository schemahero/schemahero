package mysql

import (
	"fmt"
	"strings"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RemoveForeignKeyStatement(tableName string, foreignKey *types.ForeignKey) (string, error) {
	return fmt.Sprintf("alter table %s drop constraint %s", tableName, foreignKey.Name), nil
}

func AddForeignKeyStatement(tableName string, schemaForeignKey *schemasv1alpha2.SQLTableForeignKey) (string, error) {
	return fmt.Sprintf("alter table %s add constraint %s foreign key (%s) references %s (%s)",
			tableName,
			types.GenerateFKName(tableName, schemaForeignKey),
			strings.Join(schemaForeignKey.Columns, ", "),
			schemaForeignKey.References.Table,
			strings.Join(schemaForeignKey.References.Columns, ", ")),
		nil
}
