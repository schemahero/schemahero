package postgres

import "regexp"

// oidClassRegexp matches a full-string typed literal like `'pending'::text` or
// `'{}'::jsonb`. It is anchored so embedded casts inside expressions
// (e.g. nextval('seq'::regclass)) are left unchanged.
var oidClassRegexp = regexp.MustCompile(`^'(.*)'::.+$`)

func stripOIDClass(value string) string {
	matches := oidClassRegexp.FindStringSubmatch(value)
	if len(matches) == 2 {
		return matches[1]
	}
	return value
}
