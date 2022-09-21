package types

import (
	"fmt"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

type Index struct {
	Columns  []string
	Name     string
	IsUnique bool
}

func (idx *Index) Equals(other *Index) bool {
	if idx.Name != other.Name {
		return false
	}

	if idx.IsUnique != other.IsUnique {
		return false
	}

	if len(idx.Columns) != len(other.Columns) {
		return false
	}

	for _, otherColumn := range other.Columns {
		for _, col := range idx.Columns {
			if col == otherColumn {
				goto NextColumn
			}
		}

		return false

	NextColumn:
	}

	return true
}

func IndexToMysqlSchemaIndex(index *Index) *schemasv1alpha4.MysqlTableIndex {
	schemaIndex := schemasv1alpha4.MysqlTableIndex{
		Columns:  index.Columns,
		Name:     index.Name,
		IsUnique: index.IsUnique,
	}

	return &schemaIndex
}

func IndexToPostgresqlSchemaIndex(index *Index) *schemasv1alpha4.PostgresqlTableIndex {
	schemaIndex := schemasv1alpha4.PostgresqlTableIndex{
		Columns:  index.Columns,
		Name:     index.Name,
		IsUnique: index.IsUnique,
	}

	return &schemaIndex
}

func IndexToRqliteSchemaIndex(index *Index) *schemasv1alpha4.RqliteTableIndex {
	schemaIndex := schemasv1alpha4.RqliteTableIndex{
		Columns:  index.Columns,
		Name:     index.Name,
		IsUnique: index.IsUnique,
	}

	return &schemaIndex
}

func MysqlSchemaIndexToIndex(schemaIndex *schemasv1alpha4.MysqlTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func PostgresqlSchemaIndexToIndex(schemaIndex *schemasv1alpha4.PostgresqlTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func SqliteSchemaIndexToIndex(schemaIndex *schemasv1alpha4.SqliteTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func RqliteSchemaIndexToIndex(schemaIndex *schemasv1alpha4.RqliteTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func GenerateMysqlIndexName(tableName string, schemaIndex *schemasv1alpha4.MysqlTableIndex) string {
	indexName := fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
	if len(indexName) > 64 {
		indexName = indexName[:64]
	}
	return indexName
}

func GeneratePostgresqlIndexName(tableName string, schemaIndex *schemasv1alpha4.PostgresqlTableIndex) string {
	return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
}

func GenerateSqliteIndexName(tableName string, schemaIndex *schemasv1alpha4.SqliteTableIndex) string {
	return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
}

func GenerateRqliteIndexName(tableName string, schemaIndex *schemasv1alpha4.RqliteTableIndex) string {
	return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
}
