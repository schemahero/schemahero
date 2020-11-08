package sqlite

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func RecreateTableStatements(tableName string, sqliteTableSchema *schemasv1alpha4.SqliteTableSchema) ([]string, error) {
	statements := []string{
		"begin transaction",
	}

	// to make this deterministic (and testable) generate a hash of the new schema
	b, err := json.Marshal(sqliteTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal")
	}
	sum := sha256.Sum256(b)

	tempTableName := fmt.Sprintf("%s_%x", tableName, sum)

	statements = append(statements, fmt.Sprintf("alter table `%s` rename to `%s`", tableName, tempTableName))

	createTableStatement, err := CreateTableStatements(tableName, sqliteTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate create table statements")
	}
	statements = append(statements, createTableStatement...)

	columnNames := []string{}
	for _, column := range sqliteTableSchema.Columns {
		columnNames = append(columnNames, column.Name)
	}
	statements = append(statements,
		fmt.Sprintf("insert into %s (%s) select %s from %s", tableName, strings.Join(columnNames, ", "), strings.Join(columnNames, ", "), tempTableName),
	)

	statements = append(statements, fmt.Sprintf("drop table %s", tempTableName))
	statements = append(statements, "commit")

	return statements, nil
}

func columnsMatch(col1 *types.Column, col2 *types.Column) bool {
	if col1.DataType != col2.DataType {
		return false
	}

	if col1.Charset != col2.Charset {
		return false
	}

	if col1.Collation != col2.Collation {
		return false
	}

	if col1.ColumnDefault != nil && col2.ColumnDefault == nil {
		return false
	} else if col1.ColumnDefault == nil && col2.ColumnDefault != nil {
		return false
	} else if col1.ColumnDefault != nil && col2.ColumnDefault != nil && *col1.ColumnDefault != *col2.ColumnDefault {
		return false
	}

	col1Constraints, col2Constraints := col1.Constraints, col2.Constraints
	if col1Constraints == nil {
		col1Constraints = &types.ColumnConstraints{}
	}
	if col2Constraints == nil {
		col2Constraints = &types.ColumnConstraints{}
	}

	if !types.BoolsEqual(col1Constraints.NotNull, col2Constraints.NotNull) {
		return false
	}

	col1Attributes, col2Attributes := col1.Attributes, col2.Attributes
	if col1Attributes == nil {
		col1Attributes = &types.ColumnAttributes{}
	}
	if col2Attributes == nil {
		col2Attributes = &types.ColumnAttributes{}
	}

	if !types.BoolsEqual(col1Attributes.AutoIncrement, col2Attributes.AutoIncrement) {
		return false
	}

	return true
}

func ensureColumnConstraintsNotNullTrue(column *types.Column) {
	if column.Constraints == nil {
		column.Constraints = &types.ColumnConstraints{}
	}
	column.Constraints.NotNull = &trueValue
}
