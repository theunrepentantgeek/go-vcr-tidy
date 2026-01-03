package generic

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectDeletion is an analyzer for detecting deletion of a resource via monitoring GET requests.
// It watches requests for a successful DELETE to a specific URL and spawns a MonitorDeletion analyzer to track
// the subsequent GET requests to confirm deletion.
type DetectDeletion struct{}

var _ analyzer.Interface = &DetectDeletion{}

// NewDetectDeletion creates a new DetectDeletion analyzer.
func NewDetectDeletion() *DetectDeletion {
	return &DetectDeletion{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (*DetectDeletion) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()

	if interaction.HasMethod(i, http.MethodDelete) && interaction.WasSuccessful(i) {
		// Start monitoring for deletion confirmation via GET requests.
		log.Info(
			"Found DELETE to monitor",
			"url", reqURL.String(),
		)

		monitor := NewMonitorDeletion(reqURL)

		return analyzer.Spawn(monitor), nil
	}

	// No action needed for other interactions.
	return analyzer.Result{}, nil
}
