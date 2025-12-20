package analyzer

import (
	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Interface is an abstract representation of an analyzer that processes a sequence of interactions.
type Interface interface {
	// Analyze processes another in a series of interactions.
	Analyze(log logr.Logger, i interaction.Interface) (Result, error)
}
