package interaction

import (
	"net/url"

	"github.com/google/uuid"
)

// Interface is an abstract representation of an HTTP request/response pair.
// This provides a wrapper around interactions from go-vcr that will act as an insulation layer, allowing us to adapt to
// changes in the API provided by go-vcr without requiring extensive changes here (hopefully!). It also provides seam we
// can use for testing, as well as a natural place for reusable transformations (say, if multiple analyzers want to
// deserialize JSON from the response body).
type Interface interface {
	// ID is a unique identifier for the interaction.
	ID() uuid.UUID
	// Request returns the request portion of the interaction.
	Request() Request
	// Response returns the response portion of the interaction.
	Response() Response
}

// Request is an abstract representation of an HTTP request.
type Request interface {
	// FullURL returns the full URL of the request.
	FullURL() *url.URL
	// BaseURL returns the BaseURL of the request without any parameters
	BaseURL() *url.URL
	// The HTTP method of the request, e.g. "GET", "POST", etc.
	Method() string
}

// Response is an abstract representation of an HTTP response.
type Response interface {
	// StatusCode returns the HTTP status code of the response.
	StatusCode() int
	// ResponseHeader returns the value of the specified response header.
	Header(name string) (string, bool)
	// SetHeader sets the value of the specified response header.
	// name is the name of the header to set.
	// value is the value to set the header to.
	SetHeader(name string, value string)
	// RemoveHeader removes the specified response header.
	// name is the name of the header to remove.
	RemoveHeader(name string)
	// Body returns the body of the response.
	Body() []byte
}
