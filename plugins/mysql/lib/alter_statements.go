package mysql

import (
	"fmt"
	"strings"

	"github.com/schemahero/schemahero/pkg/database/types"
)

type AlterModifyColumnStatement struct {
	TableName      string
	ExistingColumn types.Column
	Column         types.Column
}

func (s AlterModifyColumnStatement) DDL() []string {
	isAddingNotNull := false
	if s.Column.Constraints != nil && s.Column.Constraints.NotNull != nil && *s.Column.Constraints.NotNull {
		if s.ExistingColumn.Constraints == nil {
			isAddingNotNull = true
		} else if s.ExistingColumn.Constraints.NotNull == nil {
			isAddingNotNull = true
		} else if !*s.ExistingColumn.Constraints.NotNull {
			isAddingNotNull = true
		}
	}

	if isAddingNotNull {
		return s.ddlWithNotNull()
	}

	// we will go ahead and apply the new constraints here because mysql is pretty flexible about
	// letting you alter a table adding a not null constraint without applying a default
	// so this will be familiar behavior for mysql users
	return s.ddl(false)
}

// ddlIgnoringNotNull will NOT change the "nullability" of the colume
func (s AlterModifyColumnStatement) ddl(useConstraintsFromExistingColumn bool) []string {
	stmts := []string{
		fmt.Sprintf("alter table `%s` modify column `%s` %s", s.TableName, s.Column.Name, s.Column.DataType),
	}

	if s.Column.Charset != s.ExistingColumn.Charset || s.Column.Collation != s.ExistingColumn.Collation {
		stmt := ""
		if s.Column.Charset != "" {
			stmt = fmt.Sprintf("character set %s", s.Column.Charset)
		}
		if s.Column.Collation != "" {
			if len(stmt) > 0 {
				stmt = stmt + " "
			}
			stmt = fmt.Sprintf("%scollate %s", stmt, s.Column.Collation)
		}

		if len(stmt) > 0 {
			stmts = append(stmts, stmt)
		}
	}

	if useConstraintsFromExistingColumn {
		if s.ExistingColumn.Constraints != nil {
			if *s.ExistingColumn.Constraints.NotNull {
				stmts = append(stmts, "not null")
			} else {
				stmts = append(stmts, "null")
			}
		}
	} else {
		if s.Column.Constraints != nil {
			if *s.Column.Constraints.NotNull {
				stmts = append(stmts, "not null")
			} else {
				stmts = append(stmts, "null")
			}
		}
	}

	if s.Column.Attributes != nil && s.Column.Attributes.AutoIncrement != nil && *s.Column.Attributes.AutoIncrement {
		stmts = append(stmts, "auto_increment")
	}

	if s.Column.ColumnDefault != nil {
		stmts = append(stmts, fmt.Sprintf("default \"%s\"", *s.Column.ColumnDefault))
	}

	return []string{strings.Join(stmts, " ")}
}

func (s AlterModifyColumnStatement) ddlWithNotNull() []string {
	// 1. update to add the default
	// 2. update existing values with the default
	// 3. update to set not null

	statements := []string{}

	// set the default (and change any types as necessary)
	if s.Column.ColumnDefault != nil {
		if s.ExistingColumn.ColumnDefault == nil || *s.ExistingColumn.ColumnDefault != *s.Column.ColumnDefault {
			statements = append(statements, s.ddl(true)...)
		}
	}

	// update existing values
	if s.Column.ColumnDefault != nil {
		localStatement := fmt.Sprintf("update `%s` set `%s`=%q where `%s` is null",
			s.TableName,
			s.Column.Name,
			*s.Column.ColumnDefault,
			s.Column.Name)
		statements = append(statements, localStatement)
	}

	// update including the not null
	statements = append(statements, s.ddl(false)...)

	return statements
}

type AlterDropColumnStatement struct {
	TableName string
	Column    types.Column
}

func (s AlterDropColumnStatement) DDL() []string {
	return []string{
		fmt.Sprintf("alter table `%s` drop column `%s`", s.TableName, s.Column.Name),
	}
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
