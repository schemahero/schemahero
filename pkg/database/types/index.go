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

func IndexToSchemaIndex(index *Index) *schemasv1alpha4.SQLTableIndex {
	schemaIndex := schemasv1alpha4.SQLTableIndex{
		Columns:  index.Columns,
		Name:     index.Name,
		IsUnique: index.IsUnique,
	}

	return &schemaIndex
}

func SchemaIndexToIndex(schemaIndex *schemasv1alpha4.SQLTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func GenerateIndexName(tableName string, schemaIndex *schemasv1alpha4.SQLTableIndex) string {
	return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
}
