package interaction

import "net/url"

// Interface is an abstract representation of an HTTL request/response pair.
// This provides a wrapper around interactions from go-vcr that will act as an insulation layer, allowing us to adapt to
// changes in the API provided by go-vcr without requiring extensive changes here (hopefully!). It also provides seam we
// can use for testing, as well as a natural place for reusable transformations (say, if multiple analyzers want to
// deserialize JSON from the response body).
type Interface interface {
	// FullURL returns the full URL of the request.
	FullURL() url.URL
	// URL returns the URL of the request without any parameters
	URL() url.URL
	// Request returns the request object.
	Request() Request
	// Response returns the response object.
	Response() Response
}

// Request is an abstract representation of an HTTP request.
type Request interface {
	// The HTTP method of the request, e.g. "GET", "POST", etc.
	Method() string
	// Body returns the body of the request as a string.
	Body() string
}

// Response is an abstract representation of an HTTP response.
type Response interface {
	// StatusCode returns the HTTP status code of the response.
	StatusCode() int
	// Body returns the body of the response as a string.
	Body() string
}
