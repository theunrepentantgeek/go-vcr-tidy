package must

import (
	"net/url"
	"testing"
)

// ParseURL parses a raw URL string and fails the test on error.
func ParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	return parsed
}
