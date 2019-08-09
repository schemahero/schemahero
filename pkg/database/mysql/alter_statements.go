package mysql

import (
	"fmt"
	"strings"

	"github.com/schemahero/schemahero/pkg/database/types"
)

type AlterModifyColumnStatement struct {
	TableName string
	Column    types.Column
}

func (s AlterModifyColumnStatement) String() string {
	stmts := []string{
		fmt.Sprintf("alter table `%s` modify column `%s` %s", s.TableName, s.Column.Name, s.Column.DataType),
	}
	if s.Column.Constraints != nil && s.Column.Constraints.NotNull != nil {
		if *s.Column.Constraints.NotNull {
			stmts = append(stmts, "not null")
		} else {
			stmts = append(stmts, "null")
		}
	}
	if s.Column.ColumnDefault != nil {
		stmts = append(stmts, fmt.Sprintf("default \"%s\"", *s.Column.ColumnDefault))
	}
	return strings.Join(stmts, " ")
}

type AlterDropColumnStatement struct {
	TableName string
	Column    types.Column
}

func (s AlterDropColumnStatement) String() string {
	return fmt.Sprintf("alter table `%s` drop column `%s`", s.TableName, s.Column.Name)
}

type AlterRemoveConstrantStatement struct {
	TableName  string
	Constraint types.KeyConstraint
}

func (s AlterRemoveConstrantStatement) String() string {
	if s.Constraint.IsPrimary {
		return fmt.Sprintf("alter table `%s` drop primary key", s.TableName)
	}
	return fmt.Sprintf("alter table `%s` drop index `%s`", s.TableName, s.Constraint.Name)
}

type AlterAddConstrantStatement struct {
	TableName  string
	Constraint types.KeyConstraint
}

func (s AlterAddConstrantStatement) String() string {
	stmts := []string{
		fmt.Sprintf("alter table `%s` add constraint `%s`", s.TableName, s.Constraint.GenerateName(s.TableName)),
	}
	if s.Constraint.IsPrimary {
		stmts = append(stmts, "primary key")
	}
	stmts = append(stmts, fmt.Sprintf("(`%s`)", strings.Join(s.Constraint.Columns, "`, `")))
	return strings.Join(stmts, " ")
}
