package types

import (
	"fmt"
	"sort"
	"strings"
)

type KeyConstraint struct {
	Name      string
	Columns   []string
	IsPrimary bool
}

func (k *KeyConstraint) Equals(other *KeyConstraint) bool {
	if k == nil && other == nil {
		return true
	}
	if k == nil || other == nil {
		return false
	}
	if k.IsPrimary != other.IsPrimary {
		return false
	}
	if len(k.Columns) != len(other.Columns) {
		return false
	}

	// For primary keys, we don't care about the order of columns
	if k.IsPrimary {
		// Create sorted copies of the column slices
		kColumns := make([]string, len(k.Columns))
		otherColumns := make([]string, len(other.Columns))
		
		copy(kColumns, k.Columns)
		copy(otherColumns, other.Columns)
		
		sort.Strings(kColumns)
		sort.Strings(otherColumns)
		
		// Compare the sorted slices
		for i, column := range kColumns {
			if column != otherColumns[i] {
				return false
			}
		}
		return true
	}

	// For non-primary keys, order matters
	for i, column := range k.Columns {
		if column != other.Columns[i] {
			return false
		}
	}
	return true
}

func (k *KeyConstraint) GenerateName(tableName string) string {
	if k.Name != "" {
		return k.Name
	}
	if k.IsPrimary {
		return fmt.Sprintf("%s_pkey", tableName)
	}
	return fmt.Sprintf("%s_%s_key", tableName, strings.Join(k.Columns, "_"))
}
