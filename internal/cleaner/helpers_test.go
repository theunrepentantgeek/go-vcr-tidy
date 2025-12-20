package cleaner

import "net/url"

// mustParseURL parses a raw URL string and panics on error.
//
//nolint:unparam // Temporary whlie test coverage is being built out.
func mustParseURL(rawURL string) url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return *parsed
}
