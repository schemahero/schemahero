package table

import (
	"context"
	"database/sql"

	databasesv1alpha1 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha1"
	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
	"github.com/schemahero/schemahero/pkg/database/postgres"
)

func (r *ReconcileTable) deployPostgres(connection *databasesv1alpha1.PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha1.PostgresTableSchema) error {
	db, err := sql.Open("postgres", connection.URI.Value)
	if err != nil {
		return err
	}
	defer db.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.tables where table_name = $1`
	row := db.QueryRow(query, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return err
	}

	// table needs created?
	if tableExists == 0 {
		query, err := postgres.CreateTableStatement(tableName, postgresTableSchema)
		if err != nil {
			return err
		}

		_, err = db.Exec(query)
		if err != nil {
			return err
		}

		return nil
	}

	// table needs to be altered?
	query = `select
		column_name, column_default, is_nullable, data_type,
		character_maximum_length
		from information_schema.columns 
		where table_name = $1`
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

		existingColumn := postgres.Column{
			Name:       columnName,
			DataType:   dataType,
			IsNullable: isNullable == "YES",
		}
		if columnDefault.Valid {
			existingColumn.ColumnDefault = &columnDefault.String
		}
		if charMaxLength.Valid {
			existingColumn.CharMaxLength = &charMaxLength.Int64
		}

		columnStatement, err := postgres.AlterColumnStatement(tableName, postgresTableSchema.Columns, &existingColumn)
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
			statement, err := postgres.InsertColumnStatement(tableName, desiredColumn)
			if err != nil {
				return err
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	for _, alterOrDropStatement := range alterAndDropStatements {
		if _, err = db.ExecContext(context.Background(), alterOrDropStatement); err != nil {
			return err
		}
	}

	return nil
}
