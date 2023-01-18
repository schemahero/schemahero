package rqlite

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rqlite/gorqlite"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func PlanRQLiteView(url string, viewName string, rqliteViewSchema *schemasv1alpha4.NotImplementedViewSchema) ([]string, error) {
	return nil, errors.New("not implemented")
}

func PlanRqliteTable(url string, tableName string, rqliteTableSchema *schemasv1alpha4.RqliteTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	r, err := Connect(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to rqlite")
	}
	defer r.Close()

	row, err := r.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query:     "select count(1) from sqlite_master where type=? and name=?",
		Arguments: []interface{}{"table", tableName},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query from sqlite_master")
	}
	row.Next()

	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && rqliteTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && rqliteTableSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop table "%s"`, tableName),
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
		// shortcut to create it
		queries, err := CreateTableStatements(tableName, rqliteTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statements")
		}

		return append(queries, seedDataStatements...), nil
	}

	statements, err := buildStatements(r, tableName, rqliteTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, seedDataStatements...)

	return statements, nil
}

func buildStatements(r *RqliteConnection, tableName string, rqliteTableSchema *schemasv1alpha4.RqliteTableSchema) ([]string, error) {
	query := `SELECT
p.name AS col_name,
p.type AS col_type,
p.pk AS col_is_pk,
p.dflt_value AS col_default_val,
p.[notnull] AS col_is_not_null
FROM sqlite_master m
LEFT OUTER JOIN pragma_table_info((m.name)) p
WHERE m.type = 'table'
AND m.name = ?`

	rows, err := r.db.QueryOneParameterized(gorqlite.ParameterizedStatement{
		Query:     query,
		Arguments: []interface{}{tableName},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query from sqlite_master")
	}

	statements := []string{}
	existingColumns := []types.Column{}

	for rows.Next() {
		var columnName, dataType string
		var columnDefault gorqlite.NullString
		var primaryKey, notNull int

		if err := rows.Scan(&columnName, &dataType, &primaryKey, &columnDefault, &notNull); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		existingColumn := types.Column{
			Name:        columnName,
			DataType:    dataType,
			Constraints: &types.ColumnConstraints{},
			Attributes:  &types.ColumnAttributes{},
		}

		if notNull == 1 {
			existingColumn.Constraints.NotNull = &trueValue
		} else {
			existingColumn.Constraints.NotNull = &falseValue
		}

		if columnDefault.Valid {
			// sqlite stores quotes as well, strip them.
			// e.g. 'sometext' is stored as 'sometext' instead of sometext.
			v := columnDefault.String
			if len(v) > 0 && v[0] == '"' {
				v = strings.Trim(v, `"`)
			} else if len(v) > 0 && v[0] == '\'' {
				v = strings.Trim(v, `'`)
			} else if len(v) > 0 && v[0] == '`' {
				v = strings.Trim(v, "`")
			}
			existingColumn.ColumnDefault = &v
		}

		existingColumns = append(existingColumns, existingColumn)
	}

	tableNeedsRecreate, err := checkTableNeedsRecreate(r, tableName, rqliteTableSchema, existingColumns)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if table needs recreate")
	}

	if tableNeedsRecreate {
		hardWayStatements, err := RecreateTableStatements(tableName, rqliteTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create recreate table statements")
		}

		statements = append(statements, hardWayStatements...)
	} else {
		// add new columns
		for _, desiredColumn := range rqliteTableSchema.Columns {
			isColumnPresent := false
			for _, existingColumn := range existingColumns {
				if existingColumn.Name == desiredColumn.Name {
					isColumnPresent = true
					break
				}
			}
			if !isColumnPresent {
				statement, err := InsertColumnStatement(tableName, desiredColumn)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create insert column statement")
				}

				statements = append(statements, statement)
			}
		}

		// if there are unique indexes, they'll have to be removed before dropping their columns
		indexStatements, err := BuildAlterIndexStatements(r, tableName, rqliteTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build alter index statements")
		}
		statements = append(statements, indexStatements...)

		// drop removed columns
		for _, existingColumn := range existingColumns {
			isColumnPresent := false
			for _, desiredColumn := range rqliteTableSchema.Columns {
				if existingColumn.Name == desiredColumn.Name {
					isColumnPresent = true
					break
				}
			}
			if !isColumnPresent {
				statement, err := DropColumnStatement(tableName, existingColumn)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create drop column statement")
				}

				statements = append(statements, statement)
			}
		}
	}

	return statements, nil
}

func checkTableNeedsRecreate(r *RqliteConnection, tableName string, rqliteTableSchema *schemasv1alpha4.RqliteTableSchema, existingColumns []types.Column) (bool, error) {
	// check if primary keys match
	existingPrimaryKey, err := r.GetTablePrimaryKeyColumns(tableName)
	if err != nil {
		return false, errors.Wrap(err, "failed to get table primary key")
	}
	if len(existingPrimaryKey) != len(rqliteTableSchema.PrimaryKey) {
		return true, nil
	}
nextPrimaryKeyColumn:
	for _, desiredColumn := range rqliteTableSchema.PrimaryKey {
		for _, existingColumn := range existingPrimaryKey {
			if existingColumn == desiredColumn {
				continue nextPrimaryKeyColumn
			}
		}
		return true, nil
	}

	// check if foreign keys match
	existingForeignKeys, err := r.ListTableForeignKeys("", tableName)
	if err != nil {
		return false, errors.Wrap(err, "failed to list table foreign keys")
	}
	if len(existingForeignKeys) != len(rqliteTableSchema.ForeignKeys) {
		return true, nil
	}
nextForeignKey:
	for _, desiredForeignKey := range rqliteTableSchema.ForeignKeys {
		for _, existingForeignKey := range existingForeignKeys {
			if existingForeignKey.Equals(types.RqliteSchemaForeignKeyToForeignKey(desiredForeignKey)) {
				continue nextForeignKey
			}
		}
		return true, nil
	}

	// check if columns were modified (ok if added or removed)
	for _, existingColumn := range existingColumns {
		for _, desiredColumn := range rqliteTableSchema.Columns {
			if existingColumn.Name == desiredColumn.Name {
				col, err := schemaColumnToColumn(desiredColumn)
				if err != nil {
					return false, errors.Wrap(err, "failed to convert desired column")
				}
				if !columnsMatch(*col, existingColumn) {
					return true, nil
				}
				break
			}
		}
	}

	currentIndexes, err := r.ListTableIndexes("", tableName)
	if err != nil {
		return false, errors.Wrap(err, "failed to list table indexes")
	}

desiredIndexesLoop:
	for _, desiredIndex := range rqliteTableSchema.Indexes {
		if desiredIndex.Name == "" {
			desiredIndex.Name = types.GenerateRqliteIndexName(tableName, desiredIndex)
		}
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.RqliteSchemaIndexToIndex(desiredIndex)) {
				continue desiredIndexesLoop
			}
			if currentIndex.Name == desiredIndex.Name {
				// index already exists but it changed, check if it's a constraint
				isConstraint, err := r.IndexIsAConstraint(tableName, currentIndex.Name)
				if err != nil {
					return false, errors.Wrap(err, "failed to check if index is a constraint")
				}
				if isConstraint {
					// sqlite doesn't support dropping/adding constraints, so we have to recreate the table
					return true, nil
				}
			}
		}
	}

currentIndexesLoop:
	for _, currentIndex := range currentIndexes {
		for _, desiredIndex := range rqliteTableSchema.Indexes {
			if desiredIndex.Name == currentIndex.Name {
				// if index changed, we already checked if it's a constraint above
				continue currentIndexesLoop
			}
		}
		isConstraint, err := r.IndexIsAConstraint(tableName, currentIndex.Name)
		if err != nil {
			return false, errors.Wrap(err, "failed to check if index is a constraint")
		}
		if isConstraint {
			// sqlite doesn't support dropping/adding constraints, so we have to recreate the table
			return true, nil
		}
	}

	return false, nil
}

func DeployRqliteStatements(url string, statements []string) error {
	r, err := Connect(url)
	if err != nil {
		return errors.Wrap(err, "failed to connect")
	}
	defer r.db.Close()

	// execute
	if err := executeStatements(r, statements); err != nil {
		return errors.Wrap(err, "failed to execute statements")
	}

	return nil
}

func executeStatements(r *RqliteConnection, statements []string) error {
	filteredStatements := []string{}

	for _, statement := range statements {
		if statement == "" {
			continue
		}
		filteredStatements = append(filteredStatements, statement)
	}

	if len(filteredStatements) == 0 {
		return nil
	}

	fmt.Println("Executing the following statements:")
	for _, statement := range filteredStatements {
		fmt.Println(statement)
	}

	if wrs, err := r.db.Write(filteredStatements); err != nil {
		wrErrs := []error{}
		for _, wr := range wrs {
			wrErrs = append(wrErrs, wr.Err)
		}
		return fmt.Errorf("failed to write: %v: %v", err, wrErrs)
	}

	return nil
}
