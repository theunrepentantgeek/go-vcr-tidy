package generic

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorDeferredCreation is an analyzer for tracking polling for resource creation.
// It watches for an uninterrupted sequence of GET requests returning 404 (Not Found) to a specific URL,
// followed by a GET that returns a 2xx status (indicating the resource has been created).
// Once creation is confirmed, the analyzer marks itself as Finished.
// All 404 GET requests are accumulated, and when the 2xx is seen, the analyzer indicates all but
// the first and last 404 are removable.
// If any other requests to that URL are seen (e.g. a POST or PUT), or if a GET returns a non-404
// and non-2xx status code, the analyzer abandons monitoring and marks itself as Finished.
type MonitorDeferredCreation struct {
	baseURL      *url.URL
	interactions []interaction.Interface
}

var _ analyzer.Interface = (*MonitorDeferredCreation)(nil)

// NewMonitorDeferredCreation creates a new MonitorDeferredCreation analyzer.
// firstInteraction is the initial GET→404 that triggered the detector.
func NewMonitorDeferredCreation(
	firstInteraction interaction.Interface,
) *MonitorDeferredCreation {
	return &MonitorDeferredCreation{
		baseURL:      firstInteraction.Request().BaseURL(),
		interactions: []interaction.Interface{firstInteraction},
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()
	statusCode := i.Response().StatusCode()

	switch {
	case !urltool.SameBaseURL(reqURL, m.baseURL):
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil

	case interaction.HasMethod(i, http.MethodGet) && interaction.WasSuccessful(i):
		return m.creationConfirmed(log)

	case interaction.HasMethod(i, http.MethodGet) && statusCode == http.StatusNotFound:
		// Accumulate this 404 GET request.
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil

	case interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch):
		// Resource has changed, abandon monitoring.
		log.Debug(
			"Abandoning deferred creation monitor, resource changed",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
		)

		return analyzer.Finished(), nil

	default:
		// Unexpected method or status code, abandon monitoring.
		log.Debug(
			"Abandoning deferred creation monitor due to unexpected request",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
			"statusCode", statusCode,
		)

		return analyzer.Finished(), nil
	}
}

// creationConfirmed handles the confirmation of creation via a 2xx GET response.
func (m *MonitorDeferredCreation) creationConfirmed(
	log *slog.Logger,
) (analyzer.Result, error) {
	if len(m.interactions) < 3 {
		// Not enough intermediate interactions to exclude.
		log.Debug(
			"Short deferred creation monitor, nothing to exclude",
			"url", m.baseURL.String(),
		)

		return analyzer.Finished(), nil
	}

	log.Debug(
		"Long deferred creation found, excluding intermediate 404s",
		"url", m.baseURL.String(),
		"removed", len(m.interactions)-2,
	)

	excluded := m.interactions[1 : len(m.interactions)-1]

	return analyzer.FinishedWithExclusions(excluded...), nil
}
