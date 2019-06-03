package mysql

import (
	"context"
	"database/sql"
	"fmt"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
)

func DeployMysqlTable(uri string, tableName string, mysqlTableSchema *schemasv1alpha1.SQLTableSchema) error {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}
	defer db.Close()

	databaseName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return err
	}

	// determine if the table exists
	query := `select count(1) from information_schema.TABLES where TABLE_NAME = ? and TABLE_SCHEMA = ?`
	fmt.Printf("Executing query %q\n", query)
	row := db.QueryRow(query, tableName, databaseName)
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
		_, err = db.Exec(query)
		if err != nil {
			return err
		}

		return nil
	}

	// table needs to be altered?
	query = `select
		COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH
		from information_schema.COLUMNS
		where TABLE_NAME = ?`
	fmt.Printf("Executing query %q\n", query)
	rows, err := db.Query(query, tableName)
	if err != nil {
		return err
	}
	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &dataType,
			&charMaxLength); err != nil {
			return err
		}

		foundColumnNames = append(foundColumnNames, columnName)

		existingColumn := Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &ColumnConstraints{},
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

		columnStatement, err := AlterColumnStatement(tableName, mysqlTableSchema.Columns, &existingColumn)
		if err != nil {
			return err
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
				return err
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	for _, alterOrDropStatement := range alterAndDropStatements {
		fmt.Printf("Executing query %q\n", alterOrDropStatement)
		if _, err = db.ExecContext(context.Background(), alterOrDropStatement); err != nil {
			return err
		}
	}

	return nil
}
