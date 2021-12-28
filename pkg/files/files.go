//go:build !windows
// +build !windows

package files

import (
	"path/filepath"
	"strings"
)

func IsHidden(filename string) (bool, error) {
	base := filepath.Base(filename)
	if base != "." && strings.HasPrefix(base, ".") {
		return true, nil
	} else {
		return false, nil
	}
}
