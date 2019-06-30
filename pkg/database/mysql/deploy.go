package mysql

import (
	"context"
	"database/sql"
	"fmt"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func DeployMysqlTable(uri string, tableName string, mysqlTableSchema *schemasv1alpha2.SQLTableSchema) error {
	m, err := Connect(uri)
	if err != nil {
		return err
	}
	defer m.db.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.TABLES where TABLE_NAME = ? and TABLE_SCHEMA = ?`
	row := m.db.QueryRow(query, tableName, m.databaseName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return err
	}

	if tableExists == 0 {
		// shortcut to just create it
		query, err := CreateTableStatement(tableName, mysqlTableSchema)
		if err != nil {
			return err
		}

		fmt.Printf("Executing query %q\n", query)
		_, err = m.db.Exec(query)
		if err != nil {
			return err
		}

		return nil
	}

	// table needs to be altered?
	columnStatements, err := buildColumnStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(m, columnStatements); err != nil {
		return err
	}

	// foreign key changes
	foreignKeyStatements, err := buildForeignKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(m, foreignKeyStatements); err != nil {
		return err
	}

	// index changes
	indexStatements, err := buildIndexStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(m, indexStatements); err != nil {
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

func buildColumnStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	query := `select
		COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH
		from information_schema.COLUMNS
		where TABLE_NAME = ?`
	rows, err := m.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &charMaxLength); err != nil {
			return nil, err
		}

		foundColumnNames = append(foundColumnNames, columnName)

		existingColumn := types.Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &types.ColumnConstraints{},
		}

		if isNullable == "NO" {
			existingColumn.Constraints.NotNull = &trueValue
		} else {
			existingColumn.Constraints.NotNull = &falseValue
		}

		if columnDefault.Valid {
			existingColumn.ColumnDefault = &columnDefault.String
		}
		if charMaxLength.Valid {
			existingColumn.DataType = fmt.Sprintf("%s (%d)", existingColumn.DataType, charMaxLength.Int64)
		}

		columnStatement, err := AlterColumnStatement(tableName, mysqlTableSchema.PrimaryKey, mysqlTableSchema.Columns, &existingColumn)
		if err != nil {
			return nil, err
		}

		alterAndDropStatements = append(alterAndDropStatements, columnStatement)
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
				return nil, err
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	return alterAndDropStatements, nil
}

func buildForeignKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
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

func buildIndexStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	return []string{}, nil
}
