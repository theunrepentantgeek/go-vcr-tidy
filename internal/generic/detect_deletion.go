package generic

import (
	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectDeletion is an analyzer for detecting deletion of a resource via monitoring GET requests.
// It watches requests for a successful DELETE to a specific URL and spawns a MonitorDeletion analyzer to track
// the subsequent GET requests to confirm deletion.
type DetectDeletion struct {
}

var _ analyzer.Interface = &DetectDeletion{}

// NewDetectDeletion creates a new DetectDeletion analyzer.
func NewDetectDeletion() *DetectDeletion {
	return &DetectDeletion{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (d *DetectDeletion) Analyze(
	log logr.Logger,
	interaction interaction.Interface,
) (analyzer.Result, error) {
	reqURL := interaction.Request().URL()
	method := interaction.Request().Method()
	statusCode := interaction.Response().StatusCode()

	if method == "DELETE" && statusCode >= 200 && statusCode < 300 {
		// Start monitoring for deletion confirmation via GET requests.
		monitor := NewMonitorDeletion(reqURL)
		return analyzer.Spawn(monitor), nil
	}

	// No action needed for other interactions.
	return analyzer.Result{}, nil
}
