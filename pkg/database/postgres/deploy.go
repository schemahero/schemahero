package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func PlanPostgresView(uri string, viewName string, postgresViewSchema *schemasv1alpha4.NotImplementedViewSchema) ([]string, error) {
	return nil, errors.New("not implemented")
}

func PlanPostgresTable(uri string, tableName string, postgresTableSchema *schemasv1alpha4.PostgresqlTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	p, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to postgres")
	}
	defer p.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.tables where table_name = $1`
	row := p.conn.QueryRow(context.Background(), query, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && postgresTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && postgresTableSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop table %s`, pgx.Identifier{tableName}.Sanitize()),
		}, nil
	}

	seedDataStatements := []string{}
	if seedData != nil {
		seedDataStatements, err = SeedDataStatements(tableName, postgresTableSchema, seedData)
		if err != nil {
			return nil, errors.Wrap(err, "create seed data statements")
		}
	}

	if tableExists == 0 {
		// shortcut to just create it
		queries, err := CreateTableStatements(tableName, postgresTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return append(queries, seedDataStatements...), nil
	}

	statements := []string{}

	// table needs to be altered?
	columnStatements, err := BuildColumnStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statement")
	}
	statements = append(statements, columnStatements...)

	// primary key changes
	primaryKeyStatements, err := BuildPrimaryKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build primary key statements")
	}
	statements = append(statements, primaryKeyStatements...)

	// foreign key changes
	foreignKeyStatements, err := BuildForeignKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build foreign key statements")
	}
	statements = append(statements, foreignKeyStatements...)

	// index changes
	indexStatements, err := BuildIndexStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build index statements")
	}
	statements = append(statements, indexStatements...)

	statements = append(statements, seedDataStatements...)

	return statements, nil
}

func DeployPostgresStatements(uri string, statements []string) error {
	p, err := Connect(uri)
	if err != nil {
		return err
	}
	defer p.Close()

	// execute
	if err := executeStatements(p, statements); err != nil {
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
		if _, err := p.conn.Exec(context.Background(), statement); err != nil {
			return err
		}
	}

	return nil
}

func BuildColumnStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	query := `select
column_name, column_default, is_nullable, data_type, udt_name, character_maximum_length
from information_schema.columns
where table_name = $1`
	rows, err := p.conn.Query(context.Background(), query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from information_schema")
	}
	defer rows.Close()

	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, udtName, isNullable string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &udtName, &charMaxLength); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		foundColumnNames = append(foundColumnNames, columnName)

		existingColumn := types.Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &types.ColumnConstraints{},
		}

		if dataType == "ARRAY" {
			existingColumn.IsArray = true
			existingColumn.DataType = UDTNameToDataType(udtName)
		}

		if isNullable == "NO" {
			existingColumn.Constraints.NotNull = &trueValue
		} else {
			existingColumn.Constraints.NotNull = &falseValue
		}

		if columnDefault.Valid {
			value := stripOIDClass(columnDefault.String)
			existingColumn.ColumnDefault = &value
		}
		if charMaxLength.Valid {
			existingColumn.DataType = fmt.Sprintf("%s (%d)", existingColumn.DataType, charMaxLength.Int64)
		}

		columnStatement, err := AlterColumnStatements(tableName, postgresTableSchema.PrimaryKey, postgresTableSchema.Columns, &existingColumn)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create alter column statement")
		}

		alterAndDropStatements = append(alterAndDropStatements, columnStatement...)
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
				return nil, errors.Wrap(err, "failed to create insert column statement")
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	return alterAndDropStatements, nil
}

func BuildPrimaryKeyStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
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

func BuildForeignKeyStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	foreignKeyStatements := []string{}
	droppedKeys := []string{}
	currentForeignKeys, err := p.ListTableForeignKeys(p.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, foreignKey := range postgresTableSchema.ForeignKeys {
		var statement string
		var matchedForeignKey *types.ForeignKey
		for _, currentForeignKey := range currentForeignKeys {
			if currentForeignKey.Equals(types.PostgresqlSchemaForeignKeyToForeignKey(foreignKey)) {
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
			if currentForeignKey.Equals(types.PostgresqlSchemaForeignKeyToForeignKey(foreignKey)) {
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

func BuildIndexStatements(p *PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	indexStatements := []string{}
	droppedIndexes := []string{}
	currentIndexes, err := p.ListTableIndexes(p.databaseName, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list table indexes")
	}
	currentConstraints, err := p.ListTableConstraints(p.databaseName, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list table constraints")
	}

DesiredIndexLoop:
	for _, index := range postgresTableSchema.Indexes {
		if index.Name == "" {
			index.Name = types.GeneratePostgresqlIndexName(tableName, index)
		}

		var statement string
		var matchedIndex *types.Index
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.PostgresqlSchemaIndexToIndex(index)) {
				continue DesiredIndexLoop
			}

			if currentIndex.Name == index.Name {
				matchedIndex = currentIndex
			}
		}

		// drop and readd? pg supports a little bit of alter index we should support (rename)
		if matchedIndex != nil {
			isConstraint := false
			for _, currentConstraint := range currentConstraints {
				if matchedIndex.Name == currentConstraint {
					isConstraint = true
				}
			}

			if isConstraint {
				statement = RemoveConstraintStatement(tableName, matchedIndex)
			} else {
				statement = RemoveIndexStatement(tableName, matchedIndex)
			}
			droppedIndexes = append(droppedIndexes, matchedIndex.Name)
			indexStatements = append(indexStatements, statement)
		}

		statement = AddIndexStatement(tableName, index)
		indexStatements = append(indexStatements, statement)
	}

ExistingIndexLoop:
	for _, currentIndex := range currentIndexes {
		var statement string
		isConstraint := false

		for _, index := range postgresTableSchema.Indexes {
			if currentIndex.Equals(types.PostgresqlSchemaIndexToIndex(index)) {
				continue ExistingIndexLoop
			}
		}

		for _, droppedIndex := range droppedIndexes {
			if droppedIndex == currentIndex.Name {
				continue ExistingIndexLoop
			}
		}

		for _, currentConstraint := range currentConstraints {
			if currentIndex.Name == currentConstraint {
				isConstraint = true
			}
		}

		if isConstraint {
			statement = RemoveConstraintStatement(tableName, currentIndex)
		} else {
			statement = RemoveIndexStatement(tableName, currentIndex)
		}

		indexStatements = append(indexStatements, statement)
	}

	return indexStatements, nil
}
