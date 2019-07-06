package types

import (
	"fmt"
	"strings"

	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
)

type Index struct {
	Columns  []string
	Name     string
	IsUnique bool
}

func (idx *Index) Equals(other *Index) bool {
	// TODO

	return false
}

func IndexToSchemaIndex(index *Index) *schemasv1alpha2.SQLTableIndex {
	schemaIndex := schemasv1alpha2.SQLTableIndex{
		Columns:  index.Columns,
		Name:     index.Name,
		IsUnique: index.IsUnique,
	}

	return &schemaIndex
}

func SchemaIndexToIndex(schemaIndex *schemasv1alpha2.SQLTableIndex) *Index {
	index := Index{
		Columns:  schemaIndex.Columns,
		Name:     schemaIndex.Name,
		IsUnique: schemaIndex.IsUnique,
	}

	return &index
}

func GenerateIndexName(tableName string, schemaIndex *schemasv1alpha2.SQLTableIndex) string {
	return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(schemaIndex.Columns, "_"))
}
