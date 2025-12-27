package fake

import "net/http"

// Response implementation.
type testResponse struct {
	statusCode      int
	responseBody    string
	responseHeaders map[string][]string
}

// StatusCode returns the HTTP status code of the response.
func (r *testResponse) StatusCode() int {
	return r.statusCode
}

// Header returns the value of the specified response header.
func (r *testResponse) Header(name string) (string, bool) {
	if r.responseHeaders == nil {
		return "", false
	}

	key := http.CanonicalHeaderKey(name)

	values, ok := r.responseHeaders[key]
	if !ok || len(values) == 0 {
		return "", false
	}

	return values[0], true
}

// SetResponseHeader sets a response header value for the fake interaction.
func (r *testResponse) SetResponseHeader(name, value string) {
	key := http.CanonicalHeaderKey(name)
	r.responseHeaders[key] = []string{value}
}

func (r *testResponse) Body() []byte {
	return []byte(r.responseBody)
}
