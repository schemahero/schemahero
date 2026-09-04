package postgres

import (
	"fmt"
	"strings"
)

// formatPostgresDefaultValue returns a SQL fragment suitable for DEFAULT
// clauses and UPDATE assignments.
//
// Literal values are single-quoted. SQL expressions — function calls, casts,
// already-quoted literals, and bare keywords — are emitted unchanged so
// PostgreSQL keeps them as expressions (e.g. now() stays dynamic instead of
// freezing to a timestamp literal). See https://github.com/schemahero/schemahero/issues/1345.
func formatPostgresDefaultValue(value string) string {
	if !shouldQuotePostgresDefault(value) {
		return value
	}

	return fmt.Sprintf("'%s'", stripOIDClass(value))
}

func shouldQuotePostgresDefault(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}

	// Already a SQL string literal and/or cast (e.g. 'pending', '{}'::jsonb).
	if strings.HasPrefix(trimmed, "'") {
		return false
	}

	// Function call or parenthesized expression (e.g. now(), gen_random_uuid()).
	if strings.Contains(trimmed, "(") {
		return false
	}

	// Type cast without surrounding quotes (e.g. 0::integer).
	if strings.Contains(trimmed, "::") {
		return false
	}

	switch strings.ToUpper(trimmed) {
	case "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP", "CURRENT_USER",
		"FALSE", "LOCALTIME", "LOCALTIMESTAMP", "NULL", "SESSION_USER", "TRUE", "USER":
		return false
	}

	return true
}

// normalizePostgresDefault reduces a default to the form introspection stores
// after stripOIDClass, so spec forms like "'pending'" and "'{}'::jsonb" compare
// equal to introspected "pending" / "{}".
func normalizePostgresDefault(value string) string {
	value = strings.TrimSpace(value)
	value = stripOIDClass(value)

	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
	}

	return value
}

func postgresDefaultsEqual(a *string, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return normalizePostgresDefault(*a) == normalizePostgresDefault(*b)
}
