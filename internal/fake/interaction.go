package fake

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Interaction is a fake interaction to use during testing.
type Interaction struct {
	id              uuid.UUID
	fullURL         url.URL
	baseURL         url.URL
	method          string
	statusCode      int
	requestBody     string
	responseBody    string
	responseHeaders map[string][]string
}

var (
	_ interaction.Interface = &Interaction{}
	_ fmt.Stringer          = &Interaction{}
)

// NewInteraction creates a new fake interaction with the specified properties.
// fullURL is the full URL of the request.
// method is the HTTP method of the request.
// statusCode is the HTTP status code of the response.
func NewInteraction(
	fullURL url.URL,
	method string,
	statusCode int,
) *Interaction {
	// Fake interactions use GUIDs for IDs.
	id := uuid.New()

	// Remove all query parameters to get the base URL.
	baseURL := fullURL
	baseURL.RawQuery = ""

	return &Interaction{
		id:              id,
		fullURL:         fullURL,
		baseURL:         baseURL,
		method:          method,
		statusCode:      statusCode,
		responseHeaders: make(map[string][]string),
	}
}

// ID is a unique identifier for the interaction.
func (i *Interaction) ID() uuid.UUID {
	return i.id
}

// FullURL returns the full URL of the request.
func (i *Interaction) FullURL() url.URL {
	return i.fullURL
}

// URL returns the URL of the request without any parameters.
func (i *Interaction) URL() url.URL {
	return i.baseURL
}

// String returns a one-line representation suitable for table display.
func (i *Interaction) String() string {
	return fmt.Sprintf(
		"%-6s %3d %s",
		i.method,
		i.statusCode,
		i.baseURL.String())
}

// Method returns the HTTP method of the request.
func (r *Interaction) Method() string {
	return r.method
}

// StatusCode returns the HTTP status code of the response.
func (r *Interaction) StatusCode() int {
	return r.statusCode
}

// ResponseHeader returns the first value for the named response header, if present.
func (r *Interaction) ResponseHeader(name string) (string, bool) {
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
func (r *Interaction) SetResponseHeader(name, value string) {
	key := http.CanonicalHeaderKey(name)
	r.responseHeaders[key] = []string{value}
}
