package generic

import (
	"net/http"
	"net/url"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorDeletion is an analyzer for tracking the monitoring of a DELETE request.
// It watches requests for that URL for an uninterrupted sequence of GET requests that succeed (2xx status codes)
// followed by a GET that returns 404 (Not Found).
// Once the DELETE is confirmed, the analyzer marks itself as Finished.
// All the GET request returning 2xx are accumulated over multiple interactions, and when the 404 is seen, the analyzer
// indicates all but the first and last are removable.
// If any other requests to that URL are seen (e.g. a POST or PUT), or if a GET returns a non-2xx and non-404 status
// code, the analyzer abandons monitoring and marks itself as Finished.
type MonitorDeletion struct {
	baseURL      url.URL
	interactions []interaction.Interface
}

var _ analyzer.Interface = (*MonitorDeletion)(nil)

// NewMonitorDeletion creates a new MonitorDeletion analyzer for the specified URL.
func NewMonitorDeletion(
	baseURL url.URL,
) *MonitorDeletion {
	return &MonitorDeletion{
		baseURL: baseURL,
	}
}

// Analyze processes another interaction in the sequence.
//
//nolint:cyclomatic,revive,cyclop // Complexity is acceptable for this method
func (m *MonitorDeletion) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()

	statusCode := i.Response().StatusCode()

	switch {
	case !urltool.SameBaseURL(reqURL, m.baseURL):
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil

	case interaction.HasMethod(i, http.MethodGet) && statusCode == http.StatusNotFound:
		return m.deletionConfirmed(log)

	case interaction.HasMethod(i, http.MethodGet) && statusCode >= 200 && statusCode < 300:
		// Accumulate this successful GET request.
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil

	case interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodDelete):
		// Resource has changed, abandon monitoring.
		log.Info(
			"Abandoning DELETE monitor, resource changed",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
		)

		return analyzer.Finished(), nil

	default:
		// Unexpected method or status code, abandon monitoring.
		log.Info(
			"Abandoning DELETE monitor due to unexpected request",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
			"statusCode", statusCode,
		)

		return analyzer.Finished(), nil
	}
}

// deletionConfirmed handles the confirmation of deletion via a 404 GET response.
func (m *MonitorDeletion) deletionConfirmed(
	log logr.Logger,
) (analyzer.Result, error) {
	if len(m.interactions) < 2 {
		// No intermediate interactions to exclude.
		log.Info(
			"Short DELETE monitor, nothing to exclude",
			"url", m.baseURL.String(),
		)

		return analyzer.Finished(), nil
	}

	log.Info(
		"Long DELETE found, excluding intermediate GETs",
		"url", m.baseURL.String(),
		"removed", len(m.interactions)-2,
	)

	excluded := m.interactions[1 : len(m.interactions)-1]

	return analyzer.FinishedWithExclusions(excluded...), nil
}
