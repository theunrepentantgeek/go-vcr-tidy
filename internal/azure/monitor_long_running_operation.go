package azure

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorAzureLongRunningOperation is an analyzer for monitoring Azure long-running operations.
// After detecting a long-running operation via DetectAzureLongRunningOperation, an instance of this is spawned to track
// the operation until completion.
// It watches for GET operations to the same base URL (ignoring changes to the `t` and `c` parameters).
type MonitorAzureLongRunningOperation struct {
	operationURL *url.URL                // Base URL of the long-running operation to monitor
	interactions []interaction.Interface // an ordered list of interactions related to this operation
}

var _ analyzer.Interface = &MonitorAzureLongRunningOperation{}

func NewMonitorAzureLongRunningOperation(
	operationURL *url.URL,
) *MonitorAzureLongRunningOperation {
	return &MonitorAzureLongRunningOperation{
		operationURL: urltool.BaseURL(operationURL),
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorAzureLongRunningOperation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	if !m.isRelevantGet(i) {
		return analyzer.Result{}, nil
	}

	// Check the status of the operation
	var operation struct {
		Status string `json:"status"`
	}

	err := json.Unmarshal(i.Response().Body(), &operation)
	if err != nil {
		// Not a valid operation response; ignore
		//nolint:nilerr // Just ignore invalid responses
		return analyzer.Result{}, nil
	}

	if operation.Status == "InProgress" {
		// Record the interaction and continue
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil
	}

	// Operation is complete, check whether we have any interactions to exclude
	if len(m.interactions) <= 2 {
		// No intermediate interactions to exclude.
		log.Debug(
			"Long running operation finished quickly, nothing to exclude",
			"url", m.operationURL,
		)

		return analyzer.Finished(), nil
	}

	log.Debug(
		"Long running operation finished, excluding intermediate GETs",
		"url", m.operationURL,
		"removed", len(m.interactions)-2,
	)

	excluded := m.interactions[1 : len(m.interactions)-1]

	return analyzer.FinishedWithExclusions(excluded...), nil
}

// isRelevantGet checks whether the interaction is a GET to the operation URL.
func (m *MonitorAzureLongRunningOperation) isRelevantGet(
	i interaction.Interface,
) bool {
	// Check if the interaction is for the operation URL
	if !urltool.SameBaseURL(m.operationURL, i.Request().FullURL()) {
		return false
	}

	// Check if the interaction is a GET
	if !interaction.HasMethod(i, http.MethodGet) {
		return false
	}

	return true
}
