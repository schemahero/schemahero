package mysql

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func ValidateSchema(tableName string, mysqlTableSchema *schemasv1alpha4.MysqlTableSchema) error {
	if mysqlTableSchema == nil {
		return errors.New("mysql table schema required")
	}

	if tableName == "" {
		return errors.New("table name is required")
	}

	if len(mysqlTableSchema.Columns) == 0 {
		return errors.New("at least one column is required")
	}

	columnNames := make(map[string]bool)
	primaryKeys := make([]string, 0)

	for _, column := range mysqlTableSchema.Columns {
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

	// Validate primary keys
	if len(primaryKeys) > 0 && mysqlTableSchema.PrimaryKey != nil {
		return errors.New("primary key cannot be specified in both column constraints and table constraints")
	}

	if mysqlTableSchema.PrimaryKey != nil {
		for _, column := range mysqlTableSchema.PrimaryKey {
			if !columnNames[column] {
				return errors.Errorf("primary key column %q not found in table schema", column)
			}
		}
	}

	// Validate foreign keys
	if mysqlTableSchema.ForeignKeys != nil {
		for _, fk := range mysqlTableSchema.ForeignKeys {
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
	if mysqlTableSchema.Indexes != nil {
		for _, index := range mysqlTableSchema.Indexes {
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

func validateColumnType(column schemasv1alpha4.MysqlTableColumn) error {
	// List of common MySQL data types
	validTypes := map[string]bool{
		"bigint":     true,
		"binary":     true,
		"bit":        true,
		"blob":       true,
		"char":       true,
		"date":       true,
		"datetime":   true,
		"decimal":    true,
		"double":     true,
		"enum":       true,
		"float":      true,
		"int":        true,
		"json":       true,
		"longblob":   true,
		"longtext":   true,
		"mediumblob": true,
		"mediumint":  true,
		"mediumtext": true,
		"smallint":   true,
		"text":       true,
		"time":       true,
		"timestamp":  true,
		"tinyblob":   true,
		"tinyint":    true,
		"tinytext":   true,
		"varbinary":  true,
		"varchar":    true,
		"year":       true,
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
		case "char", "varchar", "binary", "varbinary":
			if _, err := fmt.Sscanf(params, "%d", new(int)); err != nil {
				return errors.Errorf("invalid length parameter for %q", column.Type)
			}
		case "decimal", "float", "double":
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
		case "enum":
			if !strings.Contains(params, ",") && len(params) == 0 {
				return errors.Errorf("enum type must have at least one value: %q", column.Type)
			}
		}
	}

	return nil
}

func containsAttribute(attributes []string, attribute string) bool {
	for _, attr := range attributes {
		if strings.ToLower(attr) == attribute {
			return true
		}
	}
	return false
}
