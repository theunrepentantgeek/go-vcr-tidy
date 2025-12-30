package fake

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// TestInteraction is a fake interaction to use during testing.
type TestInteraction struct {
	id       uuid.UUID
	request  testRequest
	response testResponse
}

var (
	_ interaction.Interface = &TestInteraction{}
	_ fmt.Stringer          = &TestInteraction{}
)

// Interaction creates a new fake interaction with the specified properties.
// fullURL is the full URL of the request.
// method is the HTTP method of the request.
// statusCode is the HTTP status code of the response.
func Interaction(
	fullURL url.URL,
	method string,
	statusCode int,
) *TestInteraction {
	// Remove all query parameters to get the base URL.
	baseURL := urltool.BaseURL(fullURL)

	i := &TestInteraction{
		// Fake interactions use GUIDs for IDs.
		id: uuid.New(),
	}

	i.request = testRequest{
		fullURL: fullURL,
		baseURL: *baseURL,
		method:  method,
	}

	i.response = testResponse{
		statusCode:      statusCode,
		responseHeaders: make(map[string][]string),
	}

	return i
}

// ID is a unique identifier for the interaction.
func (i *TestInteraction) ID() uuid.UUID {
	return i.id
}

// Request returns the request portion of the interaction.
func (i *TestInteraction) Request() interaction.Request { return &i.request }

// Response returns the response portion of the interaction.
func (i *TestInteraction) Response() interaction.Response { return &i.response }

// String returns a one-line representation suitable for table display.
func (i *TestInteraction) String() string {
	return fmt.Sprintf(
		"%-6s %3d %s",
		i.request.method,
		i.response.statusCode,
		i.request.baseURL.String())
}

// SetResponseBody sets the response body for the fake interaction.
func (i *TestInteraction) SetResponseBody(body string) {
	i.response.responseBody = body
}
