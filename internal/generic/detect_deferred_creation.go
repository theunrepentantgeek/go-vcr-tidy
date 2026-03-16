package generic

import (
	"log/slog"
	"net/http"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectDeferredCreation is an analyzer for detecting polling for resource creation.
// It watches for GET requests that return 404 (Not Found) and spawns a MonitorDeferredCreation
// analyzer to track the subsequent GET requests until the resource is created.
type DetectDeferredCreation struct{}

var _ analyzer.Interface = &DetectDeferredCreation{}

// NewDetectDeferredCreation creates a new DetectDeferredCreation analyzer.
func NewDetectDeferredCreation() *DetectDeferredCreation {
	return &DetectDeferredCreation{}
}

// Analyze processes another interaction in the sequence.
func (*DetectDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()

	if interaction.HasMethod(i, http.MethodGet) && i.Response().StatusCode() == http.StatusNotFound {
		log.Debug(
			"Found GET 404 to monitor for deferred creation",
			"url", reqURL.String(),
		)

		monitor := NewMonitorDeferredCreation(i)

		return analyzer.Spawn(monitor), nil
	}

	return analyzer.Result{}, nil
}
