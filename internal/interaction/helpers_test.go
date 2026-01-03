package interaction_test

import "net/url"

func mustParseURL(raw string) url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return *parsed
}
