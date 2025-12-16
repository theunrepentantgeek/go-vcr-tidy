package vcrcleaner

import (
	"net/url"

	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

type vcrInteraction struct {
	interactionID uuid.UUID
	interaction   *cassette.Interaction
}

var _ interaction.Interface = &vcrInteraction{}

func newVCRInteraction(i *cassette.Interaction) *vcrInteraction {
	return &vcrInteraction{
		interactionID: uuid.New(),
		interaction:   i,
	}
}

// ID is a unique identifier for the interaction.
func (v *vcrInteraction) ID() uuid.UUID {
	return v.interactionID
}

// FullURL returns the full URL of the request.
func (v *vcrInteraction) FullURL() url.URL {
	u, err := url.Parse(v.interaction.Request.URL)
	if err != nil {
		// If parsing fails, panic (this should never happen in normal operation).
		panic(err)
	}

	return *u
}

// URL returns the URL of the request without any parameters
func (v *vcrInteraction) URL() url.URL {
	u := v.FullURL()
	u.RawQuery = ""
	u.Fragment = ""
	return u
}

// Method returns the HTTP method of the request, e.g. "GET", "POST", etc.
func (v *vcrInteraction) Method() string {
	return v.interaction.Request.Method
}

// StatusCode returns the HTTP status code of the response.
func (v *vcrInteraction) StatusCode() int {
	return v.interaction.Response.Code
}
