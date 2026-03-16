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
//
// To avoid spawning redundant monitors for the same resource, this analyzer keeps track of
// active monitors keyed by the request base URL. Only one monitor per base URL is active
// at any time.
type DetectDeferredCreation struct {
	activeMonitors map[string]struct{}
}

var _ analyzer.Interface = &DetectDeferredCreation{}

// NewDetectDeferredCreation creates a new DetectDeferredCreation analyzer.
func NewDetectDeferredCreation() *DetectDeferredCreation {
	return &DetectDeferredCreation{
		activeMonitors: make(map[string]struct{}),
	}
}

// Analyze processes another interaction in the sequence.
func (d *DetectDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Only consider GET requests for deferred creation monitoring.
	if !interaction.HasMethod(i, http.MethodGet) {
		return analyzer.Result{}, nil
	}

	reqURL := i.Request().BaseURL()
	urlKey := reqURL.String()

	// For a 404 response, spawn a monitor only if we are not already monitoring this URL.
	if i.Response().StatusCode() == http.StatusNotFound {
		if _, exists := d.activeMonitors[urlKey]; exists {
			// Already have an active monitor for this URL; avoid spawning duplicates.
			return analyzer.Result{}, nil
		}

		log.Debug(
			"Found GET 404 to monitor for deferred creation",
			"url", urlKey,
		)

		monitor := NewMonitorDeferredCreation(i)
		d.activeMonitors[urlKey] = struct{}{}

		return analyzer.Spawn(monitor), nil
	}

	// For non-404 GET responses, clear any active monitor entry for this URL.
	delete(d.activeMonitors, urlKey)

	return analyzer.Result{}, nil
}
