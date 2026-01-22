package azure

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorAzureAsynchronousOperation is an analyzer for monitoring Azure asynchronous operations.
// After detecting an asynchronous operation via DetectAzureAsynchronousOperation, an instance of this is spawned to
// track the operation until completion.
// It watches for GET operations to the same base URL.
type MonitorAzureAsynchronousOperation struct {
	operationURL *url.URL                // Base URL of the asynchronous operation to monitor
	interactions []interaction.Interface // an ordered list of interactions related to this operation
}

var _ analyzer.Interface = &MonitorAzureAsynchronousOperation{}

func NewMonitorAzureAsynchronousOperation(
	operationURL *url.URL,
) *MonitorAzureAsynchronousOperation {
	return &MonitorAzureAsynchronousOperation{
		operationURL: urltool.BaseURL(operationURL),
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorAzureAsynchronousOperation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Check if the interaction is for the operation URL
	if !urltool.SameBaseURL(m.operationURL, i.Request().FullURL()) {
		return analyzer.Result{}, nil
	}

	// Check if the interaction is a GET
	if i.Request().Method() != http.MethodGet {
		return analyzer.Result{}, nil
	}

	// If the status is 202 Accepted, and the Location header is present,
	// the operation is still in progress, collect the interaction and continue.
	location, ok := i.Response().Header("Location")
	if ok && i.Response().StatusCode() == 202 && location != "" {
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil
	}

	// Operation is complete, check whether we have any interactions to exclude
	if len(m.interactions) <= 2 {
		// No intermediate interactions to exclude.
		log.Debug(
			"Asynchronous operation finished quickly, nothing to exclude",
			"url", m.operationURL,
		)

		return analyzer.Finished(), nil
	}

	retained := m.interactions[:headerLength]
	retained = append(retained, m.interactions[len(m.interactions)-footerLength:]...)

	// Ensure Location headers are linked correctly
	relinkLocationHeaders(retained)

	excluded := m.interactions[headerLength : len(m.interactions)-footerLength]

	log.Debug(
		"Asynchronous operation finished, excluding intermediate GETs",
		"url", m.operationURL,
		"removed", len(excluded),
	)

	return analyzer.FinishedWithExclusions(excluded...), nil
}
