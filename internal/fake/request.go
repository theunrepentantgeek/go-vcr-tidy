package fake

import "net/url"

// Request implementation.
type fakeRequest struct {
	fullURL url.URL
	baseURL url.URL
	method  string
}

// FullURL returns the full URL of the request.
func (r *fakeRequest) FullURL() url.URL {
	return r.fullURL
}

// BaseURL returns the BaseURL of the request without any parameters.
func (r *fakeRequest) BaseURL() url.URL {
	return r.baseURL
}

// Method returns the HTTP method of the request.
func (r *fakeRequest) Method() string {
	return r.method
}
