package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

// PlanPostgresDataMigration generates SQL statements for PostgreSQL data migrations
func PlanPostgresDataMigration(uri string, migrationName string, operations []schemasv1alpha4.DataMigrationOperation) ([]string, error) {
	statements := []string{}

	for i, operation := range operations {
		operationStatements, err := generateDataMigrationStatement(operation, migrationName, i)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate statement for operation %d", i)
		}
		statements = append(statements, operationStatements...)
	}

	return statements, nil
}

// generateDataMigrationStatement generates SQL for a single data migration operation
func generateDataMigrationStatement(operation schemasv1alpha4.DataMigrationOperation, migrationName string, operationIndex int) ([]string, error) {
	statements := []string{}

	// Add comment to identify the migration
	comment := fmt.Sprintf("-- Data Migration: %s (Operation %d)", migrationName, operationIndex+1)
	statements = append(statements, comment)

	if operation.StaticUpdate != nil {
		stmt, err := generateStaticUpdateStatement(*operation.StaticUpdate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate static update statement")
		}
		statements = append(statements, stmt)
	} else if operation.CalculatedUpdate != nil {
		stmt, err := generateCalculatedUpdateStatement(*operation.CalculatedUpdate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate calculated update statement")
		}
		statements = append(statements, stmt)
	} else if operation.TransformUpdate != nil {
		stmts, err := generateTransformUpdateStatements(*operation.TransformUpdate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate transform update statements")
		}
		statements = append(statements, stmts...)
	} else if operation.CustomSQL != nil {
		stmt, err := generateCustomSQLStatement(*operation.CustomSQL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate custom SQL statement")
		}
		statements = append(statements, stmt)
	} else {
		return nil, errors.New("no valid operation specified")
	}

	return statements, nil
}

// generateStaticUpdateStatement creates an UPDATE statement with static values
func generateStaticUpdateStatement(operation schemasv1alpha4.StaticUpdateOperation) (string, error) {
	if operation.Table == "" {
		return "", errors.New("table name is required for static update")
	}
	if len(operation.Set) == 0 {
		return "", errors.New("at least one column must be specified in set clause")
	}

	// Build SET clause
	setParts := []string{}
	for column, value := range operation.Set {
		// Escape column name and properly quote value
		setParts = append(setParts, fmt.Sprintf(`"%s" = %s`, column, quoteSQLValue(value)))
	}

	// Build WHERE clause
	whereClause := ""
	if operation.Where != "" {
		whereClause = fmt.Sprintf(" WHERE %s", operation.Where)
	} else if operation.Condition != "" {
		whereClause = fmt.Sprintf(" WHERE %s", operation.Condition)
	}

	statement := fmt.Sprintf(`UPDATE "%s" SET %s%s`, operation.Table, strings.Join(setParts, ", "), whereClause)
	return statement, nil
}

// generateCalculatedUpdateStatement creates an UPDATE statement with calculated values
func generateCalculatedUpdateStatement(operation schemasv1alpha4.CalculatedUpdateOperation) (string, error) {
	if operation.Table == "" {
		return "", errors.New("table name is required for calculated update")
	}
	if len(operation.Calculations) == 0 {
		return "", errors.New("at least one calculation must be specified")
	}

	// Build SET clause from calculations
	setParts := []string{}
	for _, calc := range operation.Calculations {
		if calc.Column == "" || calc.Expression == "" {
			return "", errors.New("column name and expression are required for each calculation")
		}
		setParts = append(setParts, fmt.Sprintf(`"%s" = %s`, calc.Column, calc.Expression))
	}

	// Build WHERE clause
	whereClause := ""
	if operation.Where != "" {
		whereClause = fmt.Sprintf(" WHERE %s", operation.Where)
	} else if operation.Condition != "" {
		whereClause = fmt.Sprintf(" WHERE %s", operation.Condition)
	}

	statement := fmt.Sprintf(`UPDATE "%s" SET %s%s`, operation.Table, strings.Join(setParts, ", "), whereClause)
	return statement, nil
}

// generateTransformUpdateStatements creates UPDATE statements for data transformations
func generateTransformUpdateStatements(operation schemasv1alpha4.TransformUpdateOperation) ([]string, error) {
	if operation.Table == "" {
		return nil, errors.New("table name is required for transform update")
	}
	if len(operation.Transformations) == 0 {
		return nil, errors.New("at least one transformation must be specified")
	}

	statements := []string{}

	for _, transform := range operation.Transformations {
		stmt, err := generateTransformStatement(operation.Table, transform, operation.Where, operation.Condition)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate transformation for column %s", transform.Column)
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

// generateTransformStatement creates a SQL statement for a specific transformation
func generateTransformStatement(table string, transform schemasv1alpha4.DataTransformation, where, condition string) (string, error) {
	if transform.Column == "" {
		return "", errors.New("column name is required for transformation")
	}

	var expression string
	var err error

	switch transform.TransformType {
	case "timezone_convert":
		expression, err = generateTimezoneConversionExpression(transform)
	case "type_cast":
		expression, err = generateTypeCastExpression(transform)
	case "format_change":
		expression, err = generateFormatChangeExpression(transform)
	case "string_transform":
		expression, err = generateStringTransformExpression(transform)
	default:
		return "", errors.Errorf("unsupported transformation type: %s", transform.TransformType)
	}

	if err != nil {
		return "", err
	}

	// Build WHERE clause
	whereClause := ""
	if where != "" {
		whereClause = fmt.Sprintf(" WHERE %s", where)
	} else if condition != "" {
		whereClause = fmt.Sprintf(" WHERE %s", condition)
	}

	statement := fmt.Sprintf(`UPDATE "%s" SET "%s" = %s%s`, table, transform.Column, expression, whereClause)
	return statement, nil
}

// generateTimezoneConversionExpression creates timezone conversion expression
func generateTimezoneConversionExpression(transform schemasv1alpha4.DataTransformation) (string, error) {
	fromTz := transform.FromValue
	toTz := transform.ToValue

	if fromTz == "" || toTz == "" {
		return "", errors.New("fromValue and toValue are required for timezone_convert")
	}

	// PostgreSQL timezone conversion
	expression := fmt.Sprintf(`"%s" AT TIME ZONE '%s' AT TIME ZONE '%s'`, transform.Column, fromTz, toTz)
	return expression, nil
}

// generateTypeCastExpression creates type casting expression
func generateTypeCastExpression(transform schemasv1alpha4.DataTransformation) (string, error) {
	targetType := transform.ToValue
	if targetType == "" {
		return "", errors.New("toValue (target type) is required for type_cast")
	}

	expression := fmt.Sprintf(`"%s"::%s`, transform.Column, targetType)
	return expression, nil
}

// generateFormatChangeExpression creates format change expression
func generateFormatChangeExpression(transform schemasv1alpha4.DataTransformation) (string, error) {
	format := transform.ToValue
	if format == "" {
		return "", errors.New("toValue (format) is required for format_change")
	}

	// Common format changes
	switch format {
	case "uppercase":
		return fmt.Sprintf(`UPPER("%s")`, transform.Column), nil
	case "lowercase":
		return fmt.Sprintf(`LOWER("%s")`, transform.Column), nil
	case "trim":
		return fmt.Sprintf(`TRIM("%s")`, transform.Column), nil
	default:
		// Custom format using to_char or other functions
		if strings.Contains(format, "to_char") {
			return fmt.Sprintf(`%s("%s")`, format, transform.Column), nil
		}
		return "", errors.Errorf("unsupported format change: %s", format)
	}
}

// generateStringTransformExpression creates string transformation expressions
func generateStringTransformExpression(transform schemasv1alpha4.DataTransformation) (string, error) {
	transformType := transform.Parameters["type"]
	if transformType == "" {
		return "", errors.New("type parameter is required for string_transform")
	}

	switch transformType {
	case "replace":
		oldVal := transform.Parameters["old"]
		newVal := transform.Parameters["new"]
		if oldVal == "" {
			return "", errors.New("old parameter is required for string replace")
		}
		return fmt.Sprintf(`REPLACE("%s", '%s', '%s')`, transform.Column, oldVal, newVal), nil
	case "substring":
		start := transform.Parameters["start"]
		length := transform.Parameters["length"]
		if start == "" {
			return "", errors.New("start parameter is required for substring")
		}
		if length != "" {
			return fmt.Sprintf(`SUBSTRING("%s", %s, %s)`, transform.Column, start, length), nil
		}
		return fmt.Sprintf(`SUBSTRING("%s", %s)`, transform.Column, start), nil
	default:
		return "", errors.Errorf("unsupported string transformation type: %s", transformType)
	}
}

// generateCustomSQLStatement validates and returns custom SQL
func generateCustomSQLStatement(operation schemasv1alpha4.CustomSQLOperation) (string, error) {
	if operation.SQL == "" {
		return "", errors.New("SQL statement is required for custom SQL operation")
	}

	// Basic validation if requested
	if operation.Validate {
		if err := validateCustomSQL(operation.SQL); err != nil {
			return "", errors.Wrap(err, "custom SQL validation failed")
		}
	}

	return operation.SQL, nil
}

// validateCustomSQL performs basic validation on custom SQL
func validateCustomSQL(sql string) error {
	sql = strings.ToUpper(strings.TrimSpace(sql))

	// Basic safety checks
	dangerousKeywords := []string{"DROP TABLE", "DROP DATABASE", "TRUNCATE", "DELETE FROM"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(sql, keyword) {
			return errors.Errorf("potentially dangerous SQL detected: %s", keyword)
		}
	}

	// Ensure it's a data manipulation statement
	allowedPrefixes := []string{"UPDATE", "INSERT", "WITH"}
	hasValidPrefix := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(sql, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return errors.Errorf("custom SQL must start with one of: %s", strings.Join(allowedPrefixes, ", "))
	}

	return nil
}

// quoteSQLValue properly quotes and escapes a SQL value
func quoteSQLValue(value string) string {
	// If the value looks like a SQL expression (contains parentheses, functions, etc.), don't quote it
	if strings.Contains(value, "(") || strings.Contains(value, "CURRENT_") || 
	   strings.Contains(value, "NOW()") || strings.Contains(value, "NULL") {
		return value
	}

	// Quote string literals and escape single quotes
	escaped := strings.ReplaceAll(value, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}
