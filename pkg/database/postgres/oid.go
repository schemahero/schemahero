package postgres

import "regexp"

var oidClassRegexp = regexp.MustCompile(`'(.*)'::.+`)

func stripOIDClass(value string) string {
	matches := oidClassRegexp.FindStringSubmatch(value)
	if len(matches) == 2 {
		return matches[1]
	}
	return value
}
