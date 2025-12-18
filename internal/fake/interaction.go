package fake

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Interaction is a fake interaction to use during testing.
type Interaction struct {
	id       uuid.UUID
	request  fakeRequest
	response fakeResponse
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
	// Remove all query parameters to get the base URL.
	baseURL := fullURL
	baseURL.RawQuery = ""

	i := &Interaction{
		// Fake interactions use GUIDs for IDs.
		id: uuid.New(),
	}

	i.request = fakeRequest{
		fullURL: fullURL,
		baseURL: baseURL,
		method:  method,
	}

	i.response = fakeResponse{
		statusCode:      statusCode,
		responseHeaders: make(map[string][]string),
	}

	return i
}

// ID is a unique identifier for the interaction.
func (i *Interaction) ID() uuid.UUID {
	return i.id
}

// Request returns the request portion of the interaction.
func (i *Interaction) Request() interaction.Request { return &i.request }

// Response returns the response portion of the interaction.
func (i *Interaction) Response() interaction.Response { return &i.response }

// String returns a one-line representation suitable for table display.
func (i *Interaction) String() string {
	return fmt.Sprintf(
		"%-6s %3d %s",
		i.request.method,
		i.response.statusCode,
		i.request.baseURL.String())
}
