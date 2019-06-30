package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/schemahero/schemahero/pkg/database/types"
)

var (
	trueValue  = true
	falseValue = false
)

func (p *PostgresConnection) ListTables() ([]string, error) {
	query := "select table_name from information_schema.tables where table_catalog = $1 and table_schema = $2"

	rows, err := p.db.Query(query, p.databaseName, "public")
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

func (p *PostgresConnection) ListTableIndexes(databaseName string, tableName string) ([]*types.Index, error) {
	// started with this: https://stackoverflow.com/questions/6777456/list-all-index-names-column-names-and-its-table-name-of-a-postgresql-database
	query := `select i.relname as indname,
	am.amname as indam,
	array(
	  select pg_get_indexdef(idx.indexrelid, k + 1, true)
	  from generate_subscripts(idx.indkey, 1) as k
	  order by k
	) as indkey_names
	from pg_index as idx
	join pg_class as i
	on i.oid = idx.indexrelid
	join pg_am as am
	on i.relam = am.oid
	where idx.indrelid = $1::regclass`
	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	indexes := make([]*types.Index, 0, 0)
	for rows.Next() {
		var name, columns, method string
		if err := rows.Scan(&name, &method, &columns); err != nil {
			return nil, err
		}

		fmt.Printf("name %s columns = %s, method = %s\n", name, columns, method)
	}

	return indexes, nil
}

func (p *PostgresConnection) ListTableForeignKeys(databaseName string, tableName string) ([]*types.ForeignKey, error) {
	// Starting with a query here: https://stackoverflow.com/questions/1152260/postgres-sql-to-list-table-foreign-keys
	// TODO SchemaHero implementation needs to include a schema (database) here
	// this is pg specific because composite fks need to be handled and this might be the only way?
	query := `select
	att2.attname as "child_column",
	cl.relname as "parent_table",
	att.attname as "parent_column",
  	rc.delete_rule,
	conname
    from
       (select
	    unnest(con1.conkey) as "parent",
	    unnest(con1.confkey) as "child",
	    con1.confrelid,
	    con1.conrelid,
	    con1.conname
	from
	    pg_class cl
	    join pg_namespace ns on cl.relnamespace = ns.oid
	    join pg_constraint con1 on con1.conrelid = cl.oid
	where
	    cl.relname = $1
	    and con1.contype = 'f'
       ) con
       join pg_attribute att on
	   att.attrelid = con.confrelid and att.attnum = con.child
       join pg_class cl on
	   cl.oid = con.confrelid
       join pg_attribute att2 on
	   att2.attrelid = con.conrelid and att2.attnum = con.parent
       join information_schema.referential_constraints rc on
       rc.constraint_name = conname`

	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	foreignKeys := make([]*types.ForeignKey, 0, 0)
	for rows.Next() {
		var childColumn, parentColumn, parentTable, name, deleteRule string

		if err := rows.Scan(&childColumn, &parentTable, &parentColumn, &deleteRule, &name); err != nil {
			return nil, err
		}

		foreignKey := types.ForeignKey{
			Name:          name,
			ParentTable:   parentTable,
			OnDelete:      deleteRule,
			ChildColumns:  []string{childColumn},
			ParentColumns: []string{parentColumn},
		}

		for _, foundFk := range foreignKeys {
			if foundFk.Name == name {
				foundFk.ChildColumns = append(foreignKey.ChildColumns, childColumn)
				foundFk.ParentColumns = append(foreignKey.ParentColumns, parentColumn)

				goto Appended
			}
		}

		foreignKeys = append(foreignKeys, &foreignKey)

	Appended:
	}

	return foreignKeys, nil
}

func (p *PostgresConnection) GetTablePrimaryKey(tableName string) ([]string, error) {
	// TODO we should be adding a database name on this select
	query := `select c.column_name
from information_schema.table_constraints tc
join information_schema.constraint_column_usage as ccu using (constraint_schema, constraint_name)
join information_schema.columns as c on c.table_schema = tc.constraint_schema
  and tc.table_name = c.table_name and ccu.column_name = c.column_name
where constraint_type = 'PRIMARY KEY' and tc.table_name = $1`

	rows, err := p.db.Query(query, tableName)
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

func (p *PostgresConnection) GetTableSchema(tableName string) ([]*types.Column, error) {
	query := "select column_name, data_type, character_maximum_length, column_default, is_nullable from information_schema.columns where table_name = $1"

	rows, err := p.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]*types.Column, 0, 0)
	for rows.Next() {
		column := types.Column{}

		var maxLength sql.NullInt64
		var isNullable string
		var columnDefault sql.NullString

		if err := rows.Scan(&column.Name, &column.DataType, &maxLength, &columnDefault, &isNullable); err != nil {
			return nil, err
		}

		if isNullable == "NO" {
			column.Constraints = &types.ColumnConstraints{
				NotNull: &trueValue,
			}
		} else {
			column.Constraints = &types.ColumnConstraints{
				NotNull: &falseValue,
			}
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		if maxLength.Valid {
			column.DataType = fmt.Sprintf("%s (%d)", column.DataType, maxLength.Int64)
		}

		columns = append(columns, &column)
	}

	return columns, nil
}
