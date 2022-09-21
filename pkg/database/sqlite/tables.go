package sqlite

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func (s *SqliteConnection) ListTables() ([]*types.Table, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table'"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
	}
	defer rows.Close()

	tables := []*types.Table{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		tables = append(tables, &types.Table{
			Name: tableName,
		})
	}

	return tables, nil
}

func (s *SqliteConnection) ListTableIndexes(_ string, tableName string) ([]*types.Index, error) {
	query := `
SELECT DISTINCT il.name AS index_name, il.'unique', ii.name AS column_name
FROM
	sqlite_master AS m,
	pragma_index_list(m.name) AS il,
	pragma_index_info(il.name) AS ii
WHERE m.type='table' AND m.name=? AND il.origin!='pk'
`
	rows, err := s.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query indexes")
	}
	defer rows.Close()

	indexesMap := map[string]*types.Index{}

	for rows.Next() {
		var name string
		var isUnique bool
		var column string

		if err := rows.Scan(&name, &isUnique, &column); err != nil {
			return nil, err
		}

		var index *types.Index
		if i, ok := indexesMap[name]; ok {
			index = i
		} else {
			index = &types.Index{
				Name:     name,
				IsUnique: isUnique,
			}
			indexesMap[name] = index
		}

		index.Columns = append(index.Columns, column)
	}

	indexes := []*types.Index{}
	for _, index := range indexesMap {
		indexes = append(indexes, index)
	}

	return indexes, nil
}

func (s *SqliteConnection) IndexIsAConstraint(tableName string, indexName string) (bool, error) {
	query := `
SELECT DISTINCT il.origin
FROM
	sqlite_master AS m,
	pragma_index_list(m.name) AS il
WHERE m.type='table' AND m.name=? AND il.name=? AND il.origin!='pk'
`

	var origin string
	row := s.db.QueryRow(query, tableName, indexName)
	if err := row.Scan(&origin); err != nil {
		return false, errors.Wrap(err, "failed to scan")
	}

	return origin != "c", nil
}

func (s *SqliteConnection) ListTableForeignKeys(_ string, tableName string) ([]*types.ForeignKey, error) {
	query := `SELECT id, 'from' as child_column, 'table' as parent_table, 'to' as parent_column, 'on_delete' FROM pragma_foreign_key_list(?)`
	rows, err := s.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query foreign keys")
	}
	defer rows.Close()

	foreignKeysMap := map[int]*types.ForeignKey{}

	for rows.Next() {
		var id int
		var childColumn, parentColumn, parentTable, deleteRule string

		if err := rows.Scan(&id, &childColumn, &parentTable, &parentColumn, &deleteRule); err != nil {
			return nil, err
		}

		var foreignKey *types.ForeignKey
		if fk, ok := foreignKeysMap[id]; ok {
			foreignKey = fk
		} else {
			foreignKey = &types.ForeignKey{
				Name:        "", // TODO: find a way to get the name of the foreign key
				ParentTable: parentTable,
				OnDelete:    deleteRule,
			}
			foreignKeysMap[id] = foreignKey
		}

		foreignKey.ChildColumns = append(foreignKey.ChildColumns, childColumn)
		foreignKey.ParentColumns = append(foreignKey.ParentColumns, parentColumn)
	}

	foreignKeys := []*types.ForeignKey{}
	for _, fk := range foreignKeysMap {
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, nil
}

func (s *SqliteConnection) GetTablePrimaryKey(tableName string) (*types.KeyConstraint, error) {
	query := `SELECT name FROM pragma_table_info(?) WHERE pk > 0`

	rows, err := s.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query primary keys")
	}
	defer rows.Close()

	var hasKey bool

	key := types.KeyConstraint{
		Name:      "", // TODO: find a way to get the name of the primary key
		IsPrimary: true,
	}
	for rows.Next() {
		hasKey = true
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		key.Columns = append(key.Columns, columnName)
	}
	if !hasKey {
		return nil, nil
	}

	return &key, nil
}

func (s *SqliteConnection) GetTablePrimaryKeyColumns(tableName string) ([]string, error) {
	primaryKey, err := s.GetTablePrimaryKey(tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get primary key")
	}
	if primaryKey == nil {
		return nil, nil
	}
	return primaryKey.Columns, nil
}

func (s *SqliteConnection) GetTableSchema(tableName string) ([]*types.Column, error) {
	query := "select name, type, dflt_value, notnull from pragma_table_info(?)"

	rows, err := s.db.Query(query, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query table schema")
	}
	defer rows.Close()

	columns := make([]*types.Column, 0)
	for rows.Next() {
		column := types.Column{}

		var notNull bool
		var columnDefault sql.NullString

		if err := rows.Scan(&column.Name, &column.DataType, &columnDefault, &notNull); err != nil {
			return nil, err
		}

		if notNull {
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

		columns = append(columns, &column)
	}

	return columns, nil
}
