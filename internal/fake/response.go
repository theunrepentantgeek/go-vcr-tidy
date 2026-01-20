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

// SetHeader sets the value of the specified response header.
func (r *testResponse) SetHeader(name string, value string) {
	if r.responseHeaders == nil {
		r.responseHeaders = make(map[string][]string)
	}

	key := http.CanonicalHeaderKey(name)
	r.responseHeaders[key] = []string{value}
}

// RemoveHeader removes the specified response header.
func (r *testResponse) RemoveHeader(name string) {
	if r.responseHeaders == nil {
		return
	}

	key := http.CanonicalHeaderKey(name)
	delete(r.responseHeaders, key)
}

// Body returns the body of the response.
func (r *testResponse) Body() []byte {
	return []byte(r.responseBody)
}
