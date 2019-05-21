package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var (
	trueValue  = true
	falseValue = false
)

func (pg *Postgres) ListTables() ([]string, error) {
	query := "select table_name from information_schema.tables where table_catalog = $1 and table_schema = $2"

	rows, err := pg.db.Query(query, pg.databaseName, "public")
	if err != nil {
		return nil, err
	}

	tableNames := make([]string, 0, 0)
	for rows.Next() {
		tableName := ""
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

func (pg *Postgres) GetTablePrimaryKey(tableName string) ([]string, error) {
	query := `select c.column_name
from information_schema.table_constraints tc
join information_schema.constraint_column_usage as ccu using (constraint_schema, constraint_name)
join information_schema.columns as c on c.table_schema = tc.constraint_schema
  and tc.table_name = c.table_name and ccu.column_name = c.column_name
where constraint_type = 'PRIMARY KEY' and tc.table_name = $1`

	rows, err := pg.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, 0)
	for rows.Next() {
		var columnName string

		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}

		columns = append(columns, columnName)
	}

	return columns, nil
}

func (pg *Postgres) GetTableSchema(tableName string) ([]*Column, error) {
	query := "select column_name, data_type, character_maximum_length, column_default, is_nullable from information_schema.columns where table_name = $1"

	rows, err := pg.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]*Column, 0, 0)
	for rows.Next() {
		column := Column{}

		var maxLength sql.NullInt64
		var isNullable string
		var columnDefault sql.NullString

		if err := rows.Scan(&column.Name, &column.DataType, &maxLength, &columnDefault, &isNullable); err != nil {
			return nil, err
		}

		if isNullable == "NO" {
			column.Constraints = &ColumnConstraints{
				NotNull: &trueValue,
			}
		} else {
			column.Constraints = &ColumnConstraints{
				NotNull: &falseValue,
			}
		}

		if maxLength.Valid {
			column.Constraints.MaxLength = &maxLength.Int64
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		columns = append(columns, &column)
	}

	return columns, nil
}
