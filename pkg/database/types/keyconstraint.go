package types

import (
	"fmt"
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
