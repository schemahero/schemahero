package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_sanitizeName(t *testing.T) {
	sanitizeNameTests := map[string]string{
		"two_words": "two-words",
	}

	for unsanitized, expectedSanitized := range sanitizeNameTests {
		t.Run(unsanitized, func(t *testing.T) {
			sanitized := sanitizeName(unsanitized)
			assert.Equal(t, expectedSanitized, sanitized)
		})
	}
}
