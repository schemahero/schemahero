package timescaledb

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func ValidateSchema(tableName string, timescaleTableSchema *schemasv1alpha4.TimescaleDBTableSchema) error {
	if timescaleTableSchema == nil {
		return errors.New("timescaledb table schema required")
	}

	if tableName == "" {
		return errors.New("table name is required")
	}

	if len(timescaleTableSchema.Columns) == 0 {
		return errors.New("at least one column is required")
	}

	columnNames := make(map[string]bool)
	primaryKeys := make([]string, 0)
	timeColumns := make([]string, 0)

	for _, column := range timescaleTableSchema.Columns {
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

		// Track time-based columns for hypertable validation
		if isTimeType(column.Type) {
			timeColumns = append(timeColumns, column.Name)
		}
	}

	// Validate primary keys
	if len(primaryKeys) > 0 && timescaleTableSchema.PrimaryKey != nil {
		return errors.New("primary key cannot be specified in both column constraints and table constraints")
	}

	if timescaleTableSchema.PrimaryKey != nil {
		for _, column := range timescaleTableSchema.PrimaryKey {
			if !columnNames[column] {
				return errors.Errorf("primary key column %q not found in table schema", column)
			}
		}
	}

	// Validate foreign keys
	if timescaleTableSchema.ForeignKeys != nil {
		for _, fk := range timescaleTableSchema.ForeignKeys {
			if len(fk.Columns) == 0 {
				return errors.New("foreign key must specify at least one column")
			}
			if len(fk.References.Columns) != len(fk.Columns) {
				return errors.Errorf("foreign key column count mismatch: %d columns references %d columns",
					len(fk.Columns), len(fk.References.Columns))
			}
			for _, column := range fk.Columns {
				if !columnNames[column] {
					return errors.Errorf("foreign key column %q not found in table schema", column)
				}
			}
		}
	}

	// Validate indexes
	if timescaleTableSchema.Indexes != nil {
		for _, index := range timescaleTableSchema.Indexes {
			if len(index.Columns) == 0 {
				return errors.New("index must specify at least one column")
			}
			for _, column := range index.Columns {
				if !columnNames[column] {
					return errors.Errorf("index column %q not found in table schema", column)
				}
			}
		}
	}

	return nil
}

func ValidateViewSchema(viewName string, timescaleViewSchema *schemasv1alpha4.TimescaleDBViewSchema) error {
	if timescaleViewSchema == nil {
		return errors.New("timescaledb view schema required")
	}

	if viewName == "" {
		return errors.New("view name is required")
	}

	if timescaleViewSchema.Query == "" {
		return errors.New("view query is required")
	}

	// Basic SQL query validation
	query := strings.TrimSpace(strings.ToUpper(timescaleViewSchema.Query))
	if !strings.HasPrefix(query, "SELECT") {
		return errors.New("view query must begin with SELECT")
	}

	return nil
}

func validateColumnType(column schemasv1alpha4.PostgresqlTableColumn) error {
	// List of TimescaleDB/PostgreSQL data types
	validTypes := map[string]bool{
		// Standard PostgreSQL types
		"bigint":    true,
		"boolean":   true,
		"char":      true,
		"date":      true,
		"decimal":   true,
		"double":    true,
		"integer":   true,
		"json":      true,
		"jsonb":     true,
		"numeric":   true,
		"real":      true,
		"smallint":  true,
		"text":      true,
		"time":      true,
		"timestamp": true,
		"uuid":      true,
		"varchar":   true,
		// TimescaleDB specific types
		"timestamptz": true,
		"interval":    true,
	}

	// Extract base type (remove size/precision/scale specifications)
	baseType := strings.ToLower(strings.Split(column.Type, "(")[0])
	baseType = strings.TrimSpace(baseType)

	if !validTypes[baseType] {
		return errors.Errorf("unsupported data type: %q", column.Type)
	}

	// Validate types with parameters
	if strings.Contains(column.Type, "(") {
		if !strings.HasSuffix(column.Type, ")") {
			return errors.Errorf("invalid type format: %q", column.Type)
		}

		params := strings.TrimSuffix(strings.Split(column.Type, "(")[1], ")")

		switch baseType {
		case "char", "varchar":
			if _, err := fmt.Sscanf(params, "%d", new(int)); err != nil {
				return errors.Errorf("invalid length parameter for %q", column.Type)
			}
		case "numeric", "decimal":
			var precision, scale int
			parts := strings.Split(params, ",")
			if len(parts) > 2 {
				return errors.Errorf("invalid precision/scale format for %q", column.Type)
			}
			if _, err := fmt.Sscanf(parts[0], "%d", &precision); err != nil {
				return errors.Errorf("invalid precision parameter for %q", column.Type)
			}
			if len(parts) == 2 {
				if _, err := fmt.Sscanf(parts[1], "%d", &scale); err != nil {
					return errors.Errorf("invalid scale parameter for %q", column.Type)
				}
				if scale > precision {
					return errors.Errorf("scale cannot be larger than precision in %q", column.Type)
				}
			}
		}
	}

	return nil
}

func isTimeType(columnType string) bool {
	timeTypes := map[string]bool{
		"timestamp":   true,
		"timestamptz": true,
		"date":        true,
	}

	baseType := strings.ToLower(strings.Split(columnType, "(")[0])
	baseType = strings.TrimSpace(baseType)

	return timeTypes[baseType]
}

func containsAttribute(attributes []string, attribute string) bool {
	for _, attr := range attributes {
		if strings.ToLower(attr) == attribute {
			return true
		}
	}
	return false
}
