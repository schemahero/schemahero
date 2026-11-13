package postgres

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// defaultFunctionCall matches an expression such as now() or gen_random_uuid()
	defaultFunctionCall = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?\(.*\)$`)

	// defaultKeywords are SQL value functions that are written without parentheses
	defaultKeywords = map[string]struct{}{
		"current_timestamp": {},
		"current_date":      {},
		"current_time":      {},
		"localtimestamp":    {},
		"localtime":         {},
		"current_user":      {},
		"session_user":      {},
	}
)

// isDefaultExpression returns true when a column default is a function call or SQL value function
// that must be sent to the database unquoted so it is evaluated rather than stored as a string.
func isDefaultExpression(value string) bool {
	trimmed := strings.TrimSpace(value)
	_, isKeyword := defaultKeywords[strings.ToLower(trimmed)]
	if isKeyword {
		return true
	}

	return defaultFunctionCall.MatchString(trimmed)
}

// formatDefault returns a column default as it should appear in a statement. Function calls are
// left as expressions and everything else is a quoted string literal.
func formatDefault(value string) string {
	if isDefaultExpression(value) {
		return strings.TrimSpace(value)
	}

	return fmt.Sprintf("'%s'", value)
}
