package interaction

import (
	"net/url"

	"github.com/google/uuid"
)

// Interface is an abstract representation of an HTTL request/response pair.
// This provides a wrapper around interactions from go-vcr that will act as an insulation layer, allowing us to adapt to
// changes in the API provided by go-vcr without requiring extensive changes here (hopefully!). It also provides seam we
// can use for testing, as well as a natural place for reusable transformations (say, if multiple analyzers want to
// deserialize JSON from the response body).
type Interface interface {
	// ID is a unique identifier for the interaction.
	ID() uuid.UUID
	// FullURL returns the full URL of the request.
	FullURL() url.URL
	// URL returns the URL of the request without any parameters
	URL() url.URL
	// The HTTP method of the request, e.g. "GET", "POST", etc.
	Method() string
	// StatusCode returns the HTTP status code of the response.
	StatusCode() int
	// ResponseHeader returns the value of the specified response header.
	ResponseHeader(name string) (string, bool)
}
