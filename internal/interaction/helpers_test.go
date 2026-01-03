package interaction_test

import (
	"net/url"
	"testing"
)

// mustParseURL parses a raw URL string and fails the test on error.
//
//nolint:unparam // raw is always the same value
func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	return *parsed
}
