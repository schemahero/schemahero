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

func PlanMysqlView(uri string, viewName string, mysqlViewSchema *schemasv1alpha4.NotImplementedViewSchema) ([]string, error) {
	return nil, errors.New("not implemented")
}

func PlanMysqlTable(uri string, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	m, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}
	defer m.Close()

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

	seedDataStatements := []string{}
	if seedData != nil {
		seedDataStatements, err = SeedDataStatements(tableName, seedData)
		if err != nil {
			return nil, errors.Wrap(err, "create seed data statements")
		}
	}

	if tableExists == 0 {
		// shortcut to just create it
		queries, err := CreateTableStatements(tableName, mysqlTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return append(queries, seedDataStatements...), nil
	}

	statements := []string{}

	// first, if the table charset or collation changed, add
	charsetAndCollationStatements, err := buildTableCharsetAndCollationStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build table charset and collation statements")
	}
	statements = append(statements, charsetAndCollationStatements...)

	// remove primary keys before removing columns
	removePrimaryKeyStatements, err := buildRemovePrimaryKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build remove primary key statements")
	}
	statements = append(statements, removePrimaryKeyStatements...)

	// indexes need to be removed before columns are removed
	removeIndexStatements, err := buildRemoveIndexStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build remove index statements")
	}
	statements = append(statements, removeIndexStatements...)

	// table needs to be altered?
	columnStatements, err := buildColumnStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, columnStatements...)

	// primary key changes
	addPrimaryKeyStatements, err := buildAddPrimaryKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build add primary key statements")
	}
	statements = append(statements, addPrimaryKeyStatements...)

	// foreign key changes
	foreignKeyStatements, err := buildForeignKeyStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build foreign key statements")
	}
	statements = append(statements, foreignKeyStatements...)

	// add indexes after columns are added
	addIndexStatements, err := buildAddIndexStatements(m, tableName, mysqlTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build add index statements")
	}
	statements = append(statements, addIndexStatements...)

	statements = append(statements, seedDataStatements...)

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

// buildTableCharsetAndCollationStatements will return the
// statements needed to modify a TABLE's charset or collation
// (not a column or a database, those are separate)
func buildTableCharsetAndCollationStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	query := `select
t.TABLE_COLLATION,
c.character_set_name FROM information_schema.TABLES t,
information_schema.COLLATION_CHARACTER_SET_APPLICABILITY c
WHERE c.collation_name = t.table_collation
AND t.table_schema = ?
AND t.table_name = ?;`
	row := m.db.QueryRow(query, m.databaseName, tableName)

	var existingTableCollation, existingTableCharset string
	if err := row.Scan(&existingTableCollation, &existingTableCharset); err != nil {
		return nil, errors.Wrap(err, "failed to read existing table charset and collate")
	}

	// get the default for the database also
	query = `SELECT default_character_set_name, default_collation_name FROM information_schema.SCHEMATA
WHERE schema_name = ?`
	row = m.db.QueryRow(query, m.databaseName)

	var databaseCollation, databaseCharset string
	if err := row.Scan(&databaseCharset, &databaseCollation); err != nil {
		return nil, errors.Wrap(err, "failed to read existing database charset and collate")
	}

	if mysqlTableSchema.Collation == "" && mysqlTableSchema.DefaultCharset == "" {
		if existingTableCollation == databaseCollation {
			if existingTableCharset == databaseCharset {
				return []string{}, nil
			}
		}
	}

	// fill in defaults where needed
	if mysqlTableSchema.Collation == "" {
		// If charset didn't change, don't change collation. MySQL has it set to the default for the charset already.
		if mysqlTableSchema.DefaultCharset == existingTableCharset {
			return []string{}, nil
		}

		// If default charset is set, but not collation, let the database pick correct collation automatically
		// Char sets that are aliases (like utf8) will not necessarily have a record in information_schema.character_sets,
		// so we can't always look up the correct collation
		if mysqlTableSchema.DefaultCharset == "" {
			mysqlTableSchema.Collation = databaseCollation
			mysqlTableSchema.DefaultCharset = databaseCharset
		}
	} else if mysqlTableSchema.DefaultCharset == "" {
		// here the collation must have been set, but not the charset
		// get the charset associated with the collation
		query = `select CHARACTER_SET_NAME from information_schema.collations where COLLATION_NAME = ?`
		row = m.db.QueryRow(query, mysqlTableSchema.Collation)
		var collationCharset string
		if err := row.Scan(&collationCharset); err != nil {
			return nil, errors.Wrapf(err, "failed to read charset for collation %s", mysqlTableSchema.Collation)
		}
		mysqlTableSchema.DefaultCharset = collationCharset
	}

	charsetMatches := false
	collationMatches := false

	if mysqlTableSchema.DefaultCharset == existingTableCharset {
		charsetMatches = true
	} else if mysqlTableSchema.DefaultCharset == "" && existingTableCharset == databaseCharset {
		charsetMatches = true
	}

	if mysqlTableSchema.Collation == existingTableCollation {
		collationMatches = true
	}

	if charsetMatches && collationMatches {
		return []string{}, nil
	}

	if mysqlTableSchema.Collation == "" {
		return []string{
			fmt.Sprintf("alter table %s convert to character set %s", tableName, mysqlTableSchema.DefaultCharset),
		}, nil
	} else {
		return []string{
			fmt.Sprintf("alter table %s convert to character set %s collate %s", tableName, mysqlTableSchema.DefaultCharset, mysqlTableSchema.Collation),
		}, nil
	}
}

// getDefaultCharsetAndCollationForTable will return the applied charset, collation for the specifed table
func getDefaultCharsetAndCollationForTable(m *MysqlConnection, tableName string) (string, string, error) {
	query := `SELECT default_character_set_name, default_collation_name FROM information_schema.SCHEMATA WHERE schema_name = ?`
	row := m.db.QueryRow(query, m.databaseName)

	defaultCharset := ""
	defaultCollation := ""
	if err := row.Scan(&defaultCharset, &defaultCollation); err != nil {
		return "", "", errors.Wrap(err, "scan default charset and collation")
	}

	query = `select
t.TABLE_COLLATION,
c.character_set_name FROM information_schema.TABLES t,
information_schema.COLLATION_CHARACTER_SET_APPLICABILITY c
WHERE c.collation_name = t.table_collation
AND t.table_schema = ?
AND t.table_name = ?`
	row = m.db.QueryRow(query, m.databaseName, tableName)

	tableCharset := sql.NullString{}
	tableCollation := sql.NullString{}
	if err := row.Scan(&tableCollation, &tableCharset); err != nil {
		return "", "", errors.Wrap(err, "scan table charset and collation")
	}

	mostSpecificCharset := defaultCharset
	if tableCharset.Valid {
		mostSpecificCharset = tableCharset.String
	}

	mostSpecificCollation := defaultCollation
	if tableCollation.Valid {
		mostSpecificCollation = tableCollation.String
	}

	return mostSpecificCharset, mostSpecificCollation, nil
}

func buildColumnStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	defaultCharset, defaultCollation, err := getDefaultCharsetAndCollationForTable(m, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "get default charset and collation")
	}

	query := `select
COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, EXTRA, COLUMN_TYPE, CHARACTER_MAXIMUM_LENGTH, CHARACTER_SET_NAME, COLLATION_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = ?
AND TABLE_NAME = ?`
	rows, err := m.db.Query(query, m.databaseName, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query from information_schema")
	}
	defer rows.Close()

	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable, extra string
		var columnDefault sql.NullString
		var charMaxLength sql.NullInt64
		var columnCharset, columnCollation sql.NullString

		if err := rows.Scan(&columnName, &columnDefault, &isNullable, &extra, &dataType, &charMaxLength, &columnCharset, &columnCollation); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		ignoreMaxLength := false
		if dataType == "text" || dataType == "tinytext" || dataType == "mediumtext" || dataType == "longtext" ||
			dataType == "blob" || dataType == "tinyblob" || dataType == "mediumblob" || dataType == "longblob" {

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

		charset := ""
		if columnCharset.Valid {
			charset = columnCharset.String
		}

		collation := ""
		if columnCollation.Valid {
			collation = columnCollation.String
		}

		existingColumn := types.Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &types.ColumnConstraints{},
			Attributes:  &types.ColumnAttributes{},
			Charset:     charset,
			Collation:   collation,
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

		columnStatement, err := AlterColumnStatements(tableName, mysqlTableSchema.PrimaryKey, mysqlTableSchema.Columns, &existingColumn, defaultCharset, defaultCollation)
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

func buildRemovePrimaryKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
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

	return statements, nil
}

func buildAddPrimaryKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
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
	if mysqlTableSchemaPrimaryKey != nil {
		statements = append(statements, AlterAddConstrantStatement{
			TableName:  tableName,
			Constraint: *mysqlTableSchemaPrimaryKey,
		}.String())
	}

	return statements, nil
}

func buildForeignKeyStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	foreignKeyStatements := []string{}
	currentForeignKeys, err := m.ListTableForeignKeys(m.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, foreignKey := range mysqlTableSchema.ForeignKeys {
		var statement string
		var matchedForeignKey *types.ForeignKey
		for _, currentForeignKey := range currentForeignKeys {
			if currentForeignKey.Equals(types.MysqlSchemaForeignKeyToForeignKey(foreignKey)) {
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
			if currentForeignKey.Equals(types.MysqlSchemaForeignKeyToForeignKey(foreignKey)) {
				goto NextCurrentFK
			}
		}

		statement = RemoveForeignKeyStatement(tableName, currentForeignKey)
		foreignKeyStatements = append(foreignKeyStatements, statement)

	NextCurrentFK:
	}

	return foreignKeyStatements, nil
}

func buildRemoveIndexStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	indexStatements := []string{}
	currentIndexes, err := m.ListTableIndexes(m.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, currentIndex := range currentIndexes {
		isMatch := false
		for _, desiredIndex := range mysqlTableSchema.Indexes {
			// if there's no name on the desired index,
			// generate one
			if desiredIndex.Name == "" {
				desiredIndex.Name = types.GenerateMysqlIndexName(tableName, desiredIndex)
			}
			if currentIndex.Equals(types.MysqlSchemaIndexToIndex(desiredIndex)) {
				isMatch = true
			}
		}

		if !isMatch {
			indexStatements = append(indexStatements, RemoveIndexStatement(tableName, currentIndex))
		}
	}
	return indexStatements, nil
}

func buildAddIndexStatements(m *MysqlConnection, tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) ([]string, error) {
	indexStatements := []string{}
	currentIndexes, err := m.ListTableIndexes(m.databaseName, tableName)
	if err != nil {
		return nil, err
	}

	for _, desiredIndex := range mysqlTableSchema.Indexes {
		isMatch := false
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.MysqlSchemaIndexToIndex(desiredIndex)) {
				isMatch = true
			}
		}

		if !isMatch {
			indexStatements = append(indexStatements, AddIndexStatement(tableName, desiredIndex))
		}
	}

	return indexStatements, nil
}
