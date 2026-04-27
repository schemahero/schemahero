package plugin

import (
	"errors"
	"fmt"
)

// DownloadError preserves the full download failure while exposing a short
// user-facing message for CLI output.
type DownloadError struct {
	Plugin string
	Err    error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message(), e.Err)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

func (e *DownloadError) Message() string {
	return fmt.Sprintf("Failed to download SchemaHero %s plugin. You can download this plugin ahead of time with 'schemahero plugin download %s'", e.Plugin, e.Plugin)
}

func DownloadErrorMessage(err error) (string, bool) {
	var downloadErr *DownloadError
	if errors.As(err, &downloadErr) {
		return downloadErr.Message(), true
	}

	return "", false
}
