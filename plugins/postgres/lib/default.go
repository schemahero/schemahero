package postgres

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// defaultFunctionCall matches an expression such as now() or gen_random_uuid() and captures the
	// function name without its schema
	defaultFunctionCall = regexp.MustCompile(`^(?:[A-Za-z_][A-Za-z0-9_]*\.)?([A-Za-z_][A-Za-z0-9_]*)\(.*\)$`)

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

	// textDefaultFunctions are the functions that are still treated as expressions on string
	// columns, where any other value that looks like a call is kept as a literal
	textDefaultFunctions = map[string]struct{}{
		"now":                   {},
		"clock_timestamp":       {},
		"statement_timestamp":   {},
		"transaction_timestamp": {},
		"gen_random_uuid":       {},
		"uuid_generate_v1":      {},
		"uuid_generate_v1mc":    {},
		"uuid_generate_v4":      {},
		"nextval":               {},
	}

	// textDataTypePrefixes are column types where a default is most likely a string literal
	textDataTypePrefixes = []string{"text", "character", "varchar", "char", "bpchar", "citext", "name"}
)

func isTextDataType(dataType string) bool {
	lower := strings.ToLower(strings.TrimSpace(dataType))
	for _, prefix := range textDataTypePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// isDefaultExpression returns true when a column default is a function call or SQL value function
// that must be sent to the database unquoted so it is evaluated rather than stored as a string.
// On string columns only well known functions count, so a literal such as "random(x)" stays quoted.
func isDefaultExpression(dataType string, value string) bool {
	trimmed := strings.TrimSpace(value)
	if _, isKeyword := defaultKeywords[strings.ToLower(trimmed)]; isKeyword {
		return true
	}

	matches := defaultFunctionCall.FindStringSubmatch(trimmed)
	if matches == nil {
		return false
	}

	if !isTextDataType(dataType) {
		return true
	}

	_, known := textDefaultFunctions[strings.ToLower(matches[1])]
	return known
}

// formatDefault returns a column default as it should appear in a statement. Function calls are
// left as expressions and everything else is a quoted string literal.
func formatDefault(dataType string, value string) string {
	if isDefaultExpression(dataType, value) {
		return strings.TrimSpace(value)
	}

	return fmt.Sprintf("'%s'", value)
}
