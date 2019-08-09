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

		if postgresTableSchema.Indexes != nil {
			for _, index := range postgresTableSchema.Indexes {
				createIndex := AddIndexStatement(tableName, index)

				fmt.Printf("Executing query: %q\n", createIndex)
				_, err := p.db.Exec(query)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// table needs to be altered?
	columnStatements, err := buildColumnStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(p, columnStatements); err != nil {
		return err
	}

	// primary key changes
	primaryKeyStatements, err := buildPrimaryKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(p, primaryKeyStatements); err != nil {
		return err
	}

	// foreign key changes
	foreignKeyStatements, err := buildForeignKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(p, foreignKeyStatements); err != nil {
		return err
	}

	// index changes
	indexStatements, err := buildIndexStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return err
	}
	if err := executeStatements(p, indexStatements); err != nil {
		return err
	}

	return nil
}

func executeStatements(p *PostgresConnection, statements []string) error {
	for _, statement := range statements {
		if statement == "" {
			continue
		}
		fmt.Printf("Executing query %q\n", statement)
		if _, err := p.db.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}

	return nil
}

func buildColumnStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	query := `select
		column_name, column_default, is_nullable, data_type, character_maximum_length
		from information_schema.columns
		where table_name = $1`
	rows, err := p.db.Query(query, tableName)
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

		columnStatement, err := AlterColumnStatement(tableName, postgresTableSchema.PrimaryKey, postgresTableSchema.Columns, &existingColumn)
		if err != nil {
			return nil, err
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
				return nil, err
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	return alterAndDropStatements, nil
}

func buildPrimaryKeyStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	currentPrimaryKey, err := p.GetTablePrimaryKey(tableName)
	if err != nil {
		return nil, err
	}
	var postgresTableSchemaPrimaryKey *types.KeyConstraint
	if len(postgresTableSchema.PrimaryKey) > 0 {
		postgresTableSchemaPrimaryKey = &types.KeyConstraint{
			IsPrimary: true,
			Columns:   postgresTableSchema.PrimaryKey,
		}
	}

	if postgresTableSchemaPrimaryKey.Equals(currentPrimaryKey) {
		return nil, nil
	}

	var statements []string
	if currentPrimaryKey != nil {
		statements = append(statements, RemoveConstrantStatement(tableName, currentPrimaryKey))
	}

	if postgresTableSchemaPrimaryKey != nil {
		statements = append(statements, AddConstrantStatement(tableName, postgresTableSchemaPrimaryKey))
	}

	return statements, nil
}

func buildForeignKeyStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	foreignKeyStatements := []string{}
	droppedKeys := []string{}
	currentForeignKeys, err := p.ListTableForeignKeys("", tableName)
	if err != nil {
		return nil, err
	}

	for _, foreignKey := range postgresTableSchema.ForeignKeys {
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
			droppedKeys = append(droppedKeys, matchedForeignKey.Name)
			foreignKeyStatements = append(foreignKeyStatements, statement)
		}

		statement = AddForeignKeyStatement(tableName, foreignKey)
		foreignKeyStatements = append(foreignKeyStatements, statement)

	Next:
	}

	for _, currentForeignKey := range currentForeignKeys {
		var statement string
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

		statement = RemoveForeignKeyStatement(tableName, currentForeignKey)
		foreignKeyStatements = append(foreignKeyStatements, statement)

	NextCurrentFK:
	}

	return foreignKeyStatements, nil
}

func buildIndexStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha2.SQLTableSchema) ([]string, error) {
	indexStatements := []string{}
	droppedIndexes := []string{}
	currentIndexes, err := p.ListTableIndexes("", tableName)
	if err != nil {
		return nil, err
	}

	for _, index := range postgresTableSchema.Indexes {
		var statement string
		var matchedIndex *types.Index
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.SchemaIndexToIndex(index)) {
				goto Next
			}

			matchedIndex = currentIndex
		}

		// drop and readd? pg supports a little bit of alter index we should support (rename)
		if matchedIndex != nil {
			statement = RemoveIndexStatement(tableName, matchedIndex)
			droppedIndexes = append(droppedIndexes, matchedIndex.Name)
			indexStatements = append(indexStatements, statement)
		}

		statement = AddIndexStatement(tableName, index)
		indexStatements = append(indexStatements, statement)

	Next:
	}

	for _, currentIndex := range currentIndexes {
		var statement string
		for _, index := range postgresTableSchema.Indexes {
			if currentIndex.Equals(types.SchemaIndexToIndex(index)) {
				goto NextCurrentIdx
			}
		}

		for _, droppedIndex := range droppedIndexes {
			if droppedIndex == currentIndex.Name {
				goto NextCurrentIdx
			}
		}

		statement = RemoveIndexStatement(tableName, currentIndex)
		indexStatements = append(indexStatements, statement)

	NextCurrentIdx:
	}

	return indexStatements, nil
}
