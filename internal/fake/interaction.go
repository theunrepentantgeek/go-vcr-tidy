package fake

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Interaction is a fake interaction to use during testing.
type Interaction struct {
	id           string
	fullURL      url.URL
	baseURL      url.URL
	method       string
	statusCode   int
	requestBody  string
	responseBody string
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
	id := uuid.New().String()

	// Remove all query parameters to get the base URL.
	baseURL := fullURL
	baseURL.RawQuery = ""

	return &Interaction{
		id:         id,
		fullURL:    fullURL,
		baseURL:    baseURL,
		method:     method,
		statusCode: statusCode,
	}
}

// FullURL returns the full URL of the request.
func (i *Interaction) FullURL() url.URL {
	return i.fullURL
}

// URL returns the URL of the request without any parameters.
func (i *Interaction) URL() url.URL {
	return i.baseURL
}

// Request returns the request object.
func (i *Interaction) Request() interaction.Request {
	return &fakeRequest{
		method: i.method,
		body:   i.requestBody,
	}
}

// Response returns the response object.
func (i *Interaction) Response() interaction.Response {
	return &fakeResponse{
		statusCode: i.statusCode,
		body:       i.responseBody,
	}
}

// String returns a one-line representation suitable for table display.
func (i *Interaction) String() string {
	return fmt.Sprintf(
		"%-6s %3d %s",
		i.method,
		i.statusCode,
		i.baseURL.String())
}

// fakeRequest is a fake HTTP request.
type fakeRequest struct {
	method string
	body   string
}

// Method returns the HTTP method of the request.
func (r *fakeRequest) Method() string {
	return r.method
}

// Body returns the body of the request as a string.
func (r *fakeRequest) Body() string {
	return r.body
}

// fakeResponse is a fake HTTP response.
type fakeResponse struct {
	statusCode int
	body       string
}

// StatusCode returns the HTTP status code of the response.
func (r *fakeResponse) StatusCode() int {
	return r.statusCode
}

// Body returns the body of the response as a string.
func (r *fakeResponse) Body() string {
	return r.body
}
