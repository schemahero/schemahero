package cassandra

import (
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func ValidateSchema(tableName string, cassandraTableSchema *schemasv1alpha4.CassandraTableSchema) error {
	if cassandraTableSchema == nil {
		return errors.New("cassandra table schema required")
	}

	if tableName == "" {
		return errors.New("table name is required")
	}

	if len(cassandraTableSchema.Columns) == 0 {
		return errors.New("at least one column is required")
	}

	columnNames := make(map[string]bool)
	partitionKeys := make([]string, 0)

	for _, column := range cassandraTableSchema.Columns {
		if column.Name == "" {
			return errors.New("column name is required")
		}

		if column.Type == "" {
			return errors.Errorf("column %q type is required", column.Name)
		}

		// Check for duplicate column names
		if columnNames[column.Name] {
			return errors.Errorf("duplicate column name %q", column.Name)
		}
		columnNames[column.Name] = true

		// Validate data types
		if err := validateColumnType(*column); err != nil {
			return errors.Wrapf(err, "invalid column type for %q", column.Name)
		}

	}

	// Validate partition keys (required in Cassandra)
	if len(partitionKeys) == 0 {
		return errors.New("at least one partition key is required")
	}

	return nil
}

func validateColumnType(column schemasv1alpha4.CassandraColumn) error {
	// List of Cassandra data types
	validTypes := map[string]bool{
		// Numeric types
		"bigint":   true,
		"decimal":  true,
		"double":   true,
		"float":    true,
		"int":      true,
		"smallint": true,
		"tinyint":  true,
		"varint":   true,
		// Text types
		"ascii":   true,
		"text":    true,
		"varchar": true,
		// Time/Date types
		"date":      true,
		"time":      true,
		"timestamp": true,
		// Other types
		"boolean":  true,
		"blob":     true,
		"inet":     true,
		"uuid":     true,
		"timeuuid": true,
		"counter":  true,
	}

	// Extract base type (remove collection type specifications)
	baseType := strings.ToLower(column.Type)
	isCollection := false

	// Handle collection types
	if strings.HasPrefix(baseType, "list<") || strings.HasPrefix(baseType, "set<") || strings.HasPrefix(baseType, "map<") {
		isCollection = true
		if !strings.HasSuffix(baseType, ">") {
			return errors.Errorf("invalid collection type format: %q", column.Type)
		}
	}

	if isCollection {
		return validateCollectionType(baseType)
	}

	// Validate non-collection type
	if !validTypes[baseType] {
		return errors.Errorf("unsupported data type: %q", column.Type)
	}

	return nil
}

func validateCollectionType(collectionType string) error {
	// Extract inner types for collections
	if strings.HasPrefix(collectionType, "list<") || strings.HasPrefix(collectionType, "set<") {
		innerType := strings.TrimSuffix(strings.SplitN(collectionType, "<", 2)[1], ">")
		return validateInnerType(innerType)
	}

	if strings.HasPrefix(collectionType, "map<") {
		innerTypes := strings.Split(strings.TrimSuffix(strings.SplitN(collectionType, "<", 2)[1], ">"), ",")
		if len(innerTypes) != 2 {
			return errors.Errorf("invalid map type format: %q", collectionType)
		}
		if err := validateInnerType(strings.TrimSpace(innerTypes[0])); err != nil {
			return errors.Wrap(err, "invalid map key type")
		}
		if err := validateInnerType(strings.TrimSpace(innerTypes[1])); err != nil {
			return errors.Wrap(err, "invalid map value type")
		}
	}

	return nil
}

func validateInnerType(innerType string) error {
	// List of valid inner types for collections
	validInnerTypes := map[string]bool{
		"ascii":     true,
		"bigint":    true,
		"blob":      true,
		"boolean":   true,
		"date":      true,
		"decimal":   true,
		"double":    true,
		"float":     true,
		"inet":      true,
		"int":       true,
		"smallint":  true,
		"text":      true,
		"time":      true,
		"timestamp": true,
		"timeuuid":  true,
		"tinyint":   true,
		"uuid":      true,
		"varchar":   true,
		"varint":    true,
	}

	if !validInnerTypes[strings.ToLower(innerType)] {
		return errors.Errorf("unsupported collection inner type: %q", innerType)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
