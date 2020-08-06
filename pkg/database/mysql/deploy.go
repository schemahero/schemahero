package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func PlanMysqlTable(uri string, tableName string, mysqlTableSchema *schemasv1alpha4.SQLTableSchema) ([]string, error) {
	m, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}
	defer m.db.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.TABLES where TABLE_NAME = ? and TABLE_SCHEMA = ?`
	row := m.db.QueryRow(query, tableName, m.databaseName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && mysqlTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && mysqlTableSchema.IsDeleted {
		return []string{
			fmt.Sprintf("drop table `%s`", tableName),
		}, nil
	}

	if tableExists == 0 {
		// shortcut to just create it
		query, err := CreateTableStatement(tableName, mysqlTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return []string{query}, nil
	}

	statements := []string{}

	// table needs to be altered?
	columnStatements, err := buildColumnStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, columnStatements...)

	// primary key changes
	primaryKeyStatements, err := buildPrimaryKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build primary key statements")
	}
	statements = append(statements, primaryKeyStatements...)

	// foreign key changes
	foreignKeyStatements, err := buildForeignKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build foreign key statements")
	}
	statements = append(statements, foreignKeyStatements...)

	// index changes
	indexStatements, err := buildIndexStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build index statements")
	}
	statements = append(statements, indexStatements...)

	return statements, nil
}

func DeployMysqlStatements(uri string, statements []string) error {
	m, err := Connect(uri)
	if err != nil {
		return err
	}
	defer m.db.Close()

	// execute
	if err := executeStatements(m, statements); err != nil {
		return err
	}

	return nil
}

func executeStatements(m *MysqlConnection, statements []string) error {
	for _, statement := range statements {
		if statement == "" {
			continue
		}
		fmt.Printf("Executing query %q\n", statement)
		if _, err := m.db.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}

	return nil
}

func buildColumnStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.SQLTableSchema) ([]string, error) {
	query := `select
COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, EXTRA, COLUMN_TYPE, CHARACTER_MAXIMUM_LENGTH
from information_schema.COLUMNS
where TABLE_NAME = ?`
	rows, err := m.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query from information_schema")
	}
	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable, extra string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &extra, &dataType, &charMaxLength); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		ignoreMaxLength := false
		if dataType == "text" || dataType == "tinytext" || dataType == "mediumtext" || dataType == "longtext" {
			ignoreMaxLength = true
		}

		if isParameterizedColumnType(dataType) {
			dataType, err = maybeParseParameterizedColumnType(dataType)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse parameterized column type")
			}
		} else if charMaxLength.Valid && !ignoreMaxLength {
			dataType = fmt.Sprintf("%s (%d)", dataType, charMaxLength.Int64)
		}

		foundColumnNames = append(foundColumnNames, columnName)

		existingColumn := types.Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &types.ColumnConstraints{},
			Attributes:  &types.ColumnAttributes{},
		}

		if isNullable == "NO" {
			existingColumn.Constraints.NotNull = &trueValue
		} else {
			existingColumn.Constraints.NotNull = &falseValue
		}

		if strings.Contains(extra, "auto_increment") {
			existingColumn.Attributes.AutoIncrement = &trueValue
		} else {
			existingColumn.Attributes.AutoIncrement = &falseValue
		}

		if columnDefault.Valid {
			existingColumn.ColumnDefault = &columnDefault.String
		}

		columnStatement, err := AlterColumnStatements(tableName, mysqlTableSchema.PrimaryKey, mysqlTableSchema.Columns, &existingColumn)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create alter column statement")
		}

		alterAndDropStatements = append(alterAndDropStatements, columnStatement...)
	}

	for _, desiredColumn := range mysqlTableSchema.Columns {
		isColumnPresent := false
		for _, foundColumn := range foundColumnNames {
			if foundColumn == desiredColumn.Name {
				isColumnPresent = true
			}
		}

		if !isColumnPresent {
			statement, err := InsertColumnStatement(tableName, desiredColumn)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create insert column statement")
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	return alterAndDropStatements, nil
}

func buildPrimaryKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.SQLTableSchema) ([]string, error) {
	currentPrimaryKey, err := m.GetTablePrimaryKey(tableName)
	if err != nil {
		return nil, err
	}
	var mysqlTableSchemaPrimaryKey *types.KeyConstraint
	if len(mysqlTableSchema.PrimaryKey) > 0 {
		mysqlTableSchemaPrimaryKey = &types.KeyConstraint{
			IsPrimary: true,
			Columns:   mysqlTableSchema.PrimaryKey,
		}
	}

	if mysqlTableSchemaPrimaryKey.Equals(currentPrimaryKey) {
		return nil, nil
	}

	var statements []string
	if currentPrimaryKey != nil {
		statements = append(statements, AlterRemoveConstrantStatement{
			TableName:  tableName,
			Constraint: *currentPrimaryKey,
		}.String())
	}

	if mysqlTableSchemaPrimaryKey != nil {
		statements = append(statements, AlterAddConstrantStatement{
			TableName:  tableName,
			Constraint: *mysqlTableSchemaPrimaryKey,
		}.String())
	}

	return statements, nil
}

func buildForeignKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.SQLTableSchema) ([]string, error) {
	foreignKeyStatements := []string{}
	currentForeignKeys, err := m.ListTableForeignKeys(m.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, foreignKey := range mysqlTableSchema.ForeignKeys {
		var statement string
		var matchedForeignKey *types.ForeignKey
		for _, currentForeignKey := range currentForeignKeys {
			if currentForeignKey.Equals(types.SchemaForeignKeyToForeignKey(foreignKey)) {
				goto Next
			}

			matchedForeignKey = currentForeignKey
		}

		// drop and readd?  is this always ok
		// TODO can we alter
		if matchedForeignKey != nil {
			statement = RemoveForeignKeyStatement(tableName, matchedForeignKey)
			foreignKeyStatements = append(foreignKeyStatements, statement)
		}

		statement = AddForeignKeyStatement(tableName, foreignKey)
		foreignKeyStatements = append(foreignKeyStatements, statement)

	Next:
	}

	for _, currentForeignKey := range currentForeignKeys {
		var statement string
		for _, foreignKey := range mysqlTableSchema.ForeignKeys {
			if currentForeignKey.Equals(types.SchemaForeignKeyToForeignKey(foreignKey)) {
				goto NextCurrentFK
			}
		}

		statement = RemoveForeignKeyStatement(tableName, currentForeignKey)
		foreignKeyStatements = append(foreignKeyStatements, statement)

	NextCurrentFK:
	}

	return foreignKeyStatements, nil
}

func buildIndexStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.SQLTableSchema) ([]string, error) {
	indexStatements := []string{}
	currentIndexes, err := m.ListTableIndexes(m.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, index := range mysqlTableSchema.Indexes {
		var statement string
		var matchedIndex *types.Index
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.SchemaIndexToIndex(index)) {
				goto Next
			}

			matchedIndex = currentIndex
		}

		// drop and readd?  mysql supports renaming indexes
		if matchedIndex != nil {
			statement = RemoveIndexStatement(tableName, matchedIndex)
			indexStatements = append(indexStatements, statement)
		}

		statement = AddIndexStatement(tableName, index)
		indexStatements = append(indexStatements, statement)

	Next:
	}

	for _, currentIndex := range currentIndexes {
		var statement string
		for _, index := range mysqlTableSchema.Indexes {
			if currentIndex.Equals(types.SchemaIndexToIndex(index)) {
				goto NextCurrentIdx
			}
		}

		statement = RemoveIndexStatement(tableName, currentIndex)
		indexStatements = append(indexStatements, statement)

	NextCurrentIdx:
	}

	return indexStatements, nil
}
