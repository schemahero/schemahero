package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func PlanSqliteTable(dsn string, tableName string, sqliteTableSchema *schemasv1alpha4.SqliteTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	s, err := Connect(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to sqlite")
	}
	defer s.Close()

	tableExists := 0
	row := s.db.QueryRow("select count(1) from sqlite_master where type=? and name=?", "table", tableName)
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && sqliteTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && sqliteTableSchema.IsDeleted {
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
		// shortcut to create it
		queries, err := CreateTableStatements(tableName, sqliteTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statements")
		}

		return append(queries, seedDataStatements...), nil
	}

	statements := []string{}

	// table needs to be altered?
	columnStatements, err := buildColumnStatements(s, tableName, sqliteTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, columnStatements...)

	// primary key changes
	// primaryKeyStatements, err := buildPrimaryKeyStatements(m, tableName, mysqlTableSchema)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to build primary key statements")
	// }
	// statements = append(statements, primaryKeyStatements...)

	// // foreign key changes
	// foreignKeyStatements, err := buildForeignKeyStatements(m, tableName, mysqlTableSchema)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to build foreign key statements")
	// }
	// statements = append(statements, foreignKeyStatements...)

	// // index changes
	// indexStatements, err := buildIndexStatements(m, tableName, mysqlTableSchema)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to build index statements")
	// }
	// statements = append(statements, indexStatements...)

	statements = append(statements, seedDataStatements...)

	return statements, nil
}

func buildColumnStatements(s *SqliteConnection, tableName string, sqliteTableSchema *schemasv1alpha4.SqliteTableSchema) ([]string, error) {
	tableNeedsRecreate := false

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
	rows, err := s.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query from sqlite_master")
	}
	defer rows.Close()

	alterAndDropStatements := []string{}
	foundColumnNames := []string{}
	for rows.Next() {
		var columnName, dataType string
		var columnDefault sql.NullString
		var primaryKey, notNull int

		if err := rows.Scan(&columnName, &dataType, &primaryKey, &columnDefault, &notNull); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		foundColumnNames = append(foundColumnNames, columnName)

		if isParameterizedColumnType(dataType) {
			dataType, err = maybeParseParameterizedColumnType(dataType)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse parameterized column type")
			}
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
			existingColumn.ColumnDefault = &columnDefault.String
		}

		// if the column exists in the desired also, we need to recreate if it different
		needsRecreate := false
		for _, desiredColumn := range sqliteTableSchema.Columns {
			if desiredColumn.Name == existingColumn.Name {
				colA, err := schemaColumnToColumn(desiredColumn)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert desired column")
				}
				if !columnsMatch(*colA, existingColumn) {
					needsRecreate = true
				}
			}
		}

		if needsRecreate {
			tableNeedsRecreate = true
		}
	}

	if !tableNeedsRecreate {
		for _, desiredColumn := range sqliteTableSchema.Columns {
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
	}

	if tableNeedsRecreate {
		hardWayStatements, err := RecreateTableStatements(tableName, sqliteTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create recreate table statements")
		}

		alterAndDropStatements = append(alterAndDropStatements, hardWayStatements...)
	}

	return alterAndDropStatements, nil
}

func DeploySqliteStatements(dsn string, statements []string) error {
	s, err := Connect(dsn)
	if err != nil {
		return err
	}
	defer s.db.Close()

	// execute
	if err := executeStatements(s, statements); err != nil {
		return err
	}

	return nil
}

func executeStatements(s *SqliteConnection, statements []string) error {
	for _, statement := range statements {
		if statement == "" {
			continue
		}
		fmt.Printf("Executing query %q\n", statement)
		if _, err := s.db.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}

	return nil
}
