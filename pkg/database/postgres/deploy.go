package postgres

import (
	"context"
	"database/sql"
	"fmt"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func DeployPostgresTable(uri string, tableName string, postgresTableSchema *schemasv1alpha2.SQLTableSchema) error {
	p, err := Connect(uri)
	if err != nil {
		return err
	}
	defer p.db.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.tables where table_name = $1`
	fmt.Printf("Executing query %q\n", query)
	row := p.db.QueryRow(query, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return err
	}

	if tableExists == 0 {
		// shortcut to just create it
		query, err := CreateTableStatement(tableName, postgresTableSchema)
		if err != nil {
			return err
		}

		fmt.Printf("Executing query %q\n", query)
		_, err = p.db.Exec(query)
		if err != nil {
			return err
		}

		return nil
	}

	// table needs to be altered?
	query = `select
		column_name, column_default, is_nullable, data_type, character_maximum_length
		from information_schema.columns
		where table_name = $1`
	fmt.Printf("Executing query %q\n", query)
	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return err
	}
	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &charMaxLength); err != nil {
			return err
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

		columnStatement, err := AlterColumnStatement(tableName, postgresTableSchema.Columns, &existingColumn)
		if err != nil {
			return err
		}

		alterAndDropStatements = append(alterAndDropStatements, columnStatement)
	}

	for _, desiredColumn := range postgresTableSchema.Columns {
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

	// foreign key changes
	droppedKeys := []string{}
	currentForeignKeys, err := p.ListTableForeignKeys("", tableName)
	if err != nil {
		return err
	}
	for _, foreignKey := range postgresTableSchema.ForeignKeys {
		var statement string
		var err error

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
			statement, err = RemoveForeignKeyStatement(tableName, matchedForeignKey)
			if err != nil {
				return err
			}
			droppedKeys = append(droppedKeys, matchedForeignKey.Name)
			alterAndDropStatements = append(alterAndDropStatements, statement)
		}

		statement, err = AddForeignKeyStatement(tableName, foreignKey)
		if err != nil {
			return err
		}
		alterAndDropStatements = append(alterAndDropStatements, statement)

	Next:
	}

	for _, currentForeignKey := range currentForeignKeys {
		var statement string
		var err error

		for _, foreignKey := range postgresTableSchema.ForeignKeys {
			if currentForeignKey.Equals(types.SchemaForeignKeyToForeignKey(foreignKey)) {
				goto NextCurrentFK
			}
		}

		for _, droppedKey := range droppedKeys {
			if droppedKey == currentForeignKey.Name {
				goto NextCurrentFK
			}
		}

		statement, err = RemoveForeignKeyStatement(tableName, currentForeignKey)
		if err != nil {
			return err
		}
		alterAndDropStatements = append(alterAndDropStatements, statement)

	NextCurrentFK:
	}

	for _, alterOrDropStatement := range alterAndDropStatements {
		if alterOrDropStatement == "" {
			continue
		}
		fmt.Printf("Executing query %q\n", alterOrDropStatement)
		if _, err = p.db.ExecContext(context.Background(), alterOrDropStatement); err != nil {
			return err
		}
	}

	return nil
}
