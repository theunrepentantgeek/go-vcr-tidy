package generic

import (
	"net/url"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// MonitorDeletion is an analyzer for tracking the monitoring of a DELETE request.
// It watches requests for that URL for an uninterrupted sequence of GET requests that succeeed (2xx status codes)
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
func NewMonitorDeletion(url url.URL) *MonitorDeletion {
	return &MonitorDeletion{
		baseURL: url,
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorDeletion) Analyze(log logr.Logger, interaction interaction.Interface) (analyzer.Result, error) {
	_ = log // currently unused
	reqURL := interaction.Request().URL()
	if reqURL.String() != m.baseURL.String() {
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil
	}

	method := interaction.Request().Method()
	if method == "GET" {
		statusCode := interaction.Response().StatusCode()
		if statusCode == 404 {
			// Deletion confirmed.
			// We should exclude all interactions except the first and last.
			if len(m.interactions) < 2 {
				// No intermediate interactions to exclude.
				return analyzer.Finished(), nil
			}

			excluded := m.interactions[1 : len(m.interactions)-1]
			return analyzer.FinishedWithExclusions(excluded...), nil
		}

		if statusCode == 301 || statusCode == 302 || statusCode == 303 || statusCode == 307 || statusCode == 308 {
			// Redirects are unexpected, abandon monitoring.
			return analyzer.Finished(), nil
		}

		if statusCode >= 200 && statusCode < 300 {
			// Accumulate this successful GET request.
			m.interactions = append(m.interactions, interaction)
			return analyzer.Result{}, nil
		}

		// Unexpected status code, abandon monitoring.
		return analyzer.Finished(), nil
	}

	// Some other method (e.g. POST, PUT), abandon monitoring.
	return analyzer.Finished(), nil
}
