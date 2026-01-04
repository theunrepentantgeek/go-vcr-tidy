package urltool

import "net/url"

// BaseURL returns the base URL of the provided URL, stripping query parameters and fragments.
func BaseURL(someURL *url.URL) *url.URL {
	result := *someURL
	result.RawQuery = ""
	result.Fragment = ""

	return result.ResolveReference(&url.URL{})
}

// SameBaseURL compares two URLs and returns true if they have the same base URL,
// ignoring query parameters and fragments.
// Canonical representations are used for comparison, to avoid encoding issues.
func SameBaseURL(
	left *url.URL,
	right *url.URL,
) bool {
	l := BaseURL(left).String()
	r := BaseURL(right).String()

	return l == r
}
