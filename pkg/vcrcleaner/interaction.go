package vcrcleaner

import (
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

type vcrInteraction struct {
	interactionID uuid.UUID
	interaction   *cassette.Interaction
	request       vcrRequest
	response      vcrResponse
}

var _ interaction.Interface = &vcrInteraction{}

func newVCRInteraction(i *cassette.Interaction) *vcrInteraction {
	result := &vcrInteraction{
		interactionID: uuid.New(),
		interaction:   i,
	}

	result.request = vcrRequest{parent: result}
	result.response = vcrResponse{parent: result}

	return result
}

// ID is a unique identifier for the interaction.
func (v *vcrInteraction) ID() uuid.UUID {
	return v.interactionID
}

// Request returns the request portion of the interaction.
func (v *vcrInteraction) Request() interaction.Request {
	return &v.request
}

// Response returns the response portion of the interaction.
func (v *vcrInteraction) Response() interaction.Response { return &v.response }
