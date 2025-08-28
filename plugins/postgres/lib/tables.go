package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database/types"
)

var (
	trueValue  = true
	falseValue = false
)

func (p *PostgresConnection) ListTables() ([]*types.Table, error) {
	tables := []*types.Table{}

	for _, schema := range p.schemas {
		query := "select table_name from information_schema.tables where table_catalog = $1 and table_schema = $2"

		rows, err := p.conn.Query(context.Background(), query, p.databaseName, schema)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to list tables in schema %s", schema))
		}

		for rows.Next() {
			tableName := ""
			if err := rows.Scan(&tableName); err != nil {
				rows.Close()
				return nil, errors.Wrap(err, "failed to scan row")
			}

			qualifiedName := tableName
			if schema != "public" {
				qualifiedName = fmt.Sprintf("%s.%s", schema, tableName)
			}

			tables = append(tables, &types.Table{
				Name:   qualifiedName,
				Schema: schema,
			})
		}
		rows.Close()
	}

	return tables, nil
}

func (p *PostgresConnection) ListTableConstraints(databaseName string, tableName string) ([]string, error) {
	schema := p.schema // Default to connection schema
	actualTableName := tableName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema = parts[0]
		actualTableName = parts[1]
	}

	query := `select constraint_name from information_schema.table_constraints
		where table_catalog = $1 and table_name = $2 and table_schema = $3`
	rows, err := p.conn.Query(context.Background(), query, databaseName, actualTableName, schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list constraints")
	}
	defer rows.Close()

	constraints := []string{}
	for rows.Next() {
		var constraint string
		if err := rows.Scan(&constraint); err != nil {
			return nil, errors.Wrap(err, "failed to scan constraint")
		}

		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

func (p *PostgresConnection) ListTableIndexes(databaseName string, tableName string) ([]*types.Index, error) {
	schema := p.schema // Default to connection schema
	actualTableName := tableName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema = parts[0]
		actualTableName = parts[1]
	}

	qualifiedTableName := actualTableName
	if schema != "public" {
		qualifiedTableName = fmt.Sprintf("%s.%s", schema, actualTableName)
	}

	// started with this: https://stackoverflow.com/questions/6777456/list-all-index-names-column-names-and-its-table-name-of-a-postgresql-database
	query := `select
	i.relname as indname,
	am.amname as indam,
	idx.indisunique,
	array(
	  select pg_get_indexdef(idx.indexrelid, k + 1, true)
	  from generate_subscripts(idx.indkey, 1) as k
	  order by k
	) as indkey_names
	from pg_index as idx
	join pg_class as i on i.oid = idx.indexrelid
	join pg_am as am on i.relam = am.oid
	where idx.indrelid = $1::regclass
	and idx.indisprimary = false`
	rows, err := p.conn.Query(context.Background(), query, qualifiedTableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query indexes")
	}
	defer rows.Close()

	indexes := make([]*types.Index, 0)
	for rows.Next() {
		var index types.Index
		var method string
		var columns []string
		if err := rows.Scan(&index.Name, &method, &index.IsUnique, &columns); err != nil {
			return nil, err
		}

		index.Columns = columns

		indexes = append(indexes, &index)
	}

	return indexes, nil
}

func (p *PostgresConnection) ListTableForeignKeys(databaseName string, tableName string) ([]*types.ForeignKey, error) {
	schema := p.schema // Default to connection schema
	actualTableName := tableName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema = parts[0]
		actualTableName = parts[1]
	}

	// Starting with a query here: https://stackoverflow.com/questions/1152260/postgres-sql-to-list-table-foreign-keys
	query := `select
	att2.attname as "child_column",
	cl.relname as "parent_table",
	att.attname as "parent_column",
  	rc.delete_rule,
	conname,
	ns2.nspname as "parent_schema"
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
	    and ns.nspname = $2
	    and con1.contype = 'f'
       ) con
       join pg_attribute att on
	   att.attrelid = con.confrelid and att.attnum = con.child
       join pg_class cl on
	   cl.oid = con.confrelid
       join pg_namespace ns2 on
       cl.relnamespace = ns2.oid
       join pg_attribute att2 on
	   att2.attrelid = con.conrelid and att2.attnum = con.parent
       join information_schema.referential_constraints rc on
       rc.constraint_name = conname`

	rows, err := p.conn.Query(context.Background(), query, actualTableName, schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query foreign keys")
	}
	defer rows.Close()

	foreignKeys := make([]*types.ForeignKey, 0)
	for rows.Next() {
		var childColumn, parentColumn, parentTable, name, deleteRule, parentSchema string

		if err := rows.Scan(&childColumn, &parentTable, &parentColumn, &deleteRule, &name, &parentSchema); err != nil {
			return nil, err
		}

		qualifiedParentTable := parentTable
		if parentSchema != "public" {
			qualifiedParentTable = fmt.Sprintf("%s.%s", parentSchema, parentTable)
		}

		foreignKey := types.ForeignKey{
			Name:          name,
			ParentTable:   qualifiedParentTable,
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

func (p *PostgresConnection) GetTablePrimaryKey(tableName string) (*types.KeyConstraint, error) {
	schema := p.schema // Default to connection schema
	actualTableName := tableName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema = parts[0]
		actualTableName = parts[1]
	}

	query := `SELECT tc.constraint_name, kcu.column_name
FROM information_schema.table_constraints  AS tc
JOIN information_schema.key_column_usage   AS kcu
  ON  kcu.constraint_catalog  = tc.constraint_catalog
  AND kcu.constraint_schema   = tc.constraint_schema
  AND kcu.constraint_name     = tc.constraint_name
WHERE tc.constraint_type = 'PRIMARY KEY'
  AND tc.table_name      = $1
  AND tc.table_schema    = $2
  AND tc.constraint_catalog = $3
ORDER BY kcu.ordinal_position`

	rows, err := p.conn.Query(context.Background(), query, actualTableName, schema, p.databaseName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query primary keys")
	}
	defer rows.Close()

	var hasKey bool

	key := types.KeyConstraint{
		IsPrimary: true,
	}
	for rows.Next() {
		hasKey = true

		var constraintName, columnName string

		if err := rows.Scan(&constraintName, &columnName); err != nil {
			return nil, err
		}

		key.Name = constraintName
		key.Columns = append(key.Columns, columnName)
	}
	if !hasKey {
		return nil, nil
	}

	return &key, nil
}

func (p *PostgresConnection) GetTableSchema(tableName string) ([]*types.Column, error) {
	schema := p.schema // Default to connection schema
	actualTableName := tableName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema = parts[0]
		actualTableName = parts[1]
	}

	query := "select column_name, data_type, character_maximum_length, column_default, is_nullable from information_schema.columns where table_name = $1 and table_schema = $2 and table_catalog = $3"

	rows, err := p.conn.Query(context.Background(), query, actualTableName, schema, p.databaseName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query table schema")
	}
	defer rows.Close()

	columns := make([]*types.Column, 0)
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
			value := stripOIDClass(columnDefault.String)
			column.ColumnDefault = &value
		}

		if maxLength.Valid {
			column.DataType = fmt.Sprintf("%s (%d)", column.DataType, maxLength.Int64)
		}

		columns = append(columns, &column)
	}

	return columns, nil
}
