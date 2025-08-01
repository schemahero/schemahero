package postgres

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func SeedDataStatements(tableName string, tableSchema *schemasv1alpha4.PostgresqlTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	statements := []string{}

	conflictInferenceSpec := findConflictInferenceSpec(tableName, tableSchema)
	for _, row := range seedData.Rows {
		cols := []string{}
		vals := []string{}
		updateVals := []string{}
		for _, col := range row.Columns {
			cols = append(cols, col.Column)
			updateVals = append(updateVals, fmt.Sprintf("excluded.%s", col.Column))
			if col.Value.Int != nil {
				vals = append(vals, strconv.Itoa(*col.Value.Int))
			} else if col.Value.Str != nil {
				vals = append(vals, fmt.Sprintf("'%s'", *col.Value.Str))
			}
		}

		var statement string
		if conflictInferenceSpec != "" {
			statement = fmt.Sprintf(`insert into %s (%s) values (%s) on conflict (%s) do update set (%s) = (%s)`, tableName, strings.Join(cols, ", "), strings.Join(vals, ", "), conflictInferenceSpec, strings.Join(cols, ", "), strings.Join(updateVals, ", "))
		} else {
			statement = fmt.Sprintf(`insert into %s (%s) values (%s)`, tableName, strings.Join(cols, ", "), strings.Join(vals, ", "))
		}
		statements = append(statements, statement)
	}

	return statements, nil
}

func SeedDataStatementsWithExistingSchema(uri string, tableName string, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// Connect to the database to retrieve existing schema
	p, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to postgres")
	}
	defer p.Close()

	// Check if table exists
	tableExists, err := CheckIfTableExists(p, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if table exists")
	}

	if !tableExists {
		return nil, errors.Errorf("table %s does not exist, cannot apply seed data without schema", tableName)
	}

	// Get primary key to determine conflict inference spec
	primaryKey, err := p.GetTablePrimaryKey(tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get table primary key")
	}

	// For seed data, we need to create a minimal PostgresqlTableSchema with just the primary key
	// since that's what the conflict inference spec calculation needs
	tableSchema := &schemasv1alpha4.PostgresqlTableSchema{}
	if primaryKey != nil {
		tableSchema.PrimaryKey = primaryKey.Columns
	}

	return SeedDataStatements(tableName, tableSchema, seedData)
}

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.PostgresqlTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnFields, err := columnAsInsert(desiredColumn)
		if err != nil {
			return nil, err
		}
		columns = append(columns, columnFields)
	}

	if len(tableSchema.PrimaryKey) > 0 {
		primaryKeyColumns := []string{}
		for _, primaryKeyColumn := range tableSchema.PrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, pgx.Identifier{primaryKeyColumn}.Sanitize())
		}

		columns = append(columns, fmt.Sprintf("primary key (%s)", strings.Join(primaryKeyColumns, ", ")))
	}

	if len(tableSchema.Indexes) > 0 {
		for _, index := range tableSchema.Indexes {
			if index.IsUnique {
				uniqueColumns := []string{}
				for _, indexColumn := range index.Columns {
					uniqueColumns = append(uniqueColumns, pgx.Identifier{indexColumn}.Sanitize())
				}
				columns = append(columns, fmt.Sprintf("constraint %q unique (%s)", types.GeneratePostgresqlIndexName(tableName, index), strings.Join(uniqueColumns, ", ")))
			}
		}
	}

	if tableSchema.ForeignKeys != nil {
		for _, foreignKey := range tableSchema.ForeignKeys {
			columns = append(columns, foreignKeyConstraintClause(tableName, foreignKey))
		}
	}

	qualifiedTableName := tableName
	if tableSchema.Schema != "" && tableSchema.Schema != "public" {
		qualifiedTableName = fmt.Sprintf("%s.%s", tableSchema.Schema, tableName)
	}

	queries := []string{
		fmt.Sprintf(`create table %s (%s)`, pgx.Identifier{qualifiedTableName}.Sanitize(), strings.Join(columns, ", ")),
	}

	var triggers []*v1alpha4.PostgresqlTableTrigger
	if tableSchema.JSONTriggers == nil || len(tableSchema.JSONTriggers) == 0 {
		triggers = tableSchema.Triggers
	} else {
		triggers = tableSchema.JSONTriggers
	}

	// Add any triggers that are defined
	for _, trigger := range triggers {
		statement, err := triggerCreateStatement(trigger, qualifiedTableName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create trigger statement")
		}

		queries = append(queries, statement)
	}
	return queries, nil
}

func findConflictInferenceSpec(tableName string, tableSchema *schemasv1alpha4.PostgresqlTableSchema) string {
	if len(tableSchema.PrimaryKey) > 0 {
		primaryKeyColumns := []string{}
		for _, primaryKeyColumn := range tableSchema.PrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, pgx.Identifier{primaryKeyColumn}.Sanitize())
		}

		return strings.Join(primaryKeyColumns, ", ")
	}

	if len(tableSchema.Indexes) > 0 {
		for _, index := range tableSchema.Indexes {
			if index.IsUnique {
				uniqueColumns := []string{}
				for _, indexColumn := range index.Columns {
					uniqueColumns = append(uniqueColumns, pgx.Identifier{indexColumn}.Sanitize())
				}
				return types.GeneratePostgresqlIndexName(tableName, index)
			}
		}
	}

	return ""
}
