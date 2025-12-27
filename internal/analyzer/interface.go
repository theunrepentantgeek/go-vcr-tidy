package analyzer

import (
	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Interface is an abstract representation of an analyzer that processes a sequence of interactions.
// Analyzers fall into two categories:
// 'Detectors' that watch for specific patterns of interactions that might indicate the start of an HTTP conversation
// that can be abbreviated. Detectors are typically created up front and run for the entire lifetime of the analysis.
// 'Monitors' that track a specific HTTP conversation, assessing whether it is suitable for abbreviation. Monitors are
// typically created dynamically by Detectors when they identify a potential conversation to monitor. A Monitior will
// terminate either when the full conversation has been analyzed, or as soon as it determines that abbreviation is not
// possible.
type Interface interface {
	// Analyze processes another in a series of interactions.
	Analyze(log logr.Logger, i interaction.Interface) (Result, error)
}
