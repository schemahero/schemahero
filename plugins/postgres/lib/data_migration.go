package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

// GenerateDataMigrationStatements generates UPDATE statements for data migrations
func GenerateDataMigrationStatements(tableName string, migrations []schemasv1alpha4.DataMigration) ([]string, error) {
	statements := []string{}

	for _, migration := range migrations {
		switch migration.Type {
		case "static":
			stmt, err := generateStaticMigration(tableName, migration)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate static migration for column %s", migration.Column)
			}
			statements = append(statements, stmt)

		case "calculated":
			stmt, err := generateCalculatedMigration(tableName, migration)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate calculated migration for column %s", migration.Column)
			}
			statements = append(statements, stmt)

		case "typeConversion":
			stmts, err := generateTypeConversionMigration(tableName, migration)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate type conversion migration for column %s", migration.Column)
			}
			statements = append(statements, stmts...)

		default:
			return nil, fmt.Errorf("unsupported migration type: %s", migration.Type)
		}
	}

	return statements, nil
}

func generateStaticMigration(tableName string, migration schemasv1alpha4.DataMigration) (string, error) {
	if migration.Value == nil {
		return "", fmt.Errorf("value is required for static migration")
	}

	// Escape single quotes in the value
	value := strings.ReplaceAll(*migration.Value, "'", "''")

	return fmt.Sprintf(
		"update %s set %s = '%s'",
		pgx.Identifier{tableName}.Sanitize(),
		pgx.Identifier{migration.Column}.Sanitize(),
		value,
	), nil
}

func generateCalculatedMigration(tableName string, migration schemasv1alpha4.DataMigration) (string, error) {
	if migration.Expression == nil {
		return "", fmt.Errorf("expression is required for calculated migration")
	}

	// Parse the expression to ensure column references are properly quoted
	expression := sanitizeExpression(*migration.Expression)

	return fmt.Sprintf(
		"update %s set %s = %s",
		pgx.Identifier{tableName}.Sanitize(),
		pgx.Identifier{migration.Column}.Sanitize(),
		expression,
	), nil
}

func generateTypeConversionMigration(tableName string, migration schemasv1alpha4.DataMigration) ([]string, error) {
	if migration.TypeConvert == nil {
		return nil, fmt.Errorf("typeConvert is required for type conversion migration")
	}

	statements := []string{}

	// First, alter the column type
	alterStmt := fmt.Sprintf(
		"alter table %s alter column %s type %s",
		pgx.Identifier{tableName}.Sanitize(),
		pgx.Identifier{migration.Column}.Sanitize(),
		migration.TypeConvert.To,
	)
	statements = append(statements, alterStmt)

	// For timestamp to timestamptz conversion, update the data
	if isTimestampConversion(migration.TypeConvert.From, migration.TypeConvert.To) {
		updateStmt := fmt.Sprintf(
			"update %s set %s = %s at time zone 'UTC'",
			pgx.Identifier{tableName}.Sanitize(),
			pgx.Identifier{migration.Column}.Sanitize(),
			pgx.Identifier{migration.Column}.Sanitize(),
		)
		statements = append(statements, updateStmt)
	}

	return statements, nil
}

func sanitizeExpression(expression string) string {
	// Simple implementation - in production, this would need proper SQL parsing
	// For now, wrap column names in quotes
	parts := strings.Fields(expression)
	result := []string{}

	for _, part := range parts {
		// Check if it's an operator or numeric value
		if isOperator(part) || isNumeric(part) {
			result = append(result, part)
		} else {
			// Assume it's a column name and quote it
			result = append(result, pgx.Identifier{part}.Sanitize())
		}
	}

	return strings.Join(result, " ")
}

func isOperator(s string) bool {
	operators := []string{"+", "-", "*", "/", "(", ")"}
	for _, op := range operators {
		if s == op {
			return true
		}
	}
	return false
}

func isNumeric(s string) bool {
	// Simple check - could be enhanced
	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}
	return true
}

func isTimestampConversion(from, to string) bool {
	fromLower := strings.ToLower(strings.TrimSpace(from))
	toLower := strings.ToLower(strings.TrimSpace(to))

	return fromLower == "timestamp" &&
		(toLower == "timestamp with time zone" || toLower == "timestamptz")
}
