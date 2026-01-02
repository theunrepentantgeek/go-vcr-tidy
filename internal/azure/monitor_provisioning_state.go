package azure

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorProvisioningState is an analyzer for monitoring Azure resource provisioning states.
// It watches for GET requests to a specific URL and tracks interactions where the provisioningState
// matches the target state (case-insensitive). When the provisioningState transitions to a
// different value, the monitor finishes and excludes all but the first and last accumulated interactions.
type MonitorProvisioningState struct {
	baseURL      *url.URL                // Base URL of the resource to monitor
	targetState  string                  // State to monitor (e.g., "Creating" or "Updating")
	interactions []interaction.Interface // Accumulated interactions with matching provisioningState
}

var _ analyzer.Interface = (*MonitorProvisioningState)(nil)

// NewMonitorProvisioningState creates a new MonitorProvisioningState analyzer.
// baseURL is the base URL of the resource to monitor.
// targetState is the provisioningState value to watch for (case-insensitive).
func NewMonitorProvisioningState(
	baseURL *url.URL,
	targetState string,
) *MonitorProvisioningState {
	return &MonitorProvisioningState{
		baseURL:     baseURL,
		targetState: targetState,
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorProvisioningState) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()
	statusCode := i.Response().StatusCode()

	switch {
	case !urltool.SameBaseURL(reqURL, m.baseURL):
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil

	case !interaction.HasMethod(i, http.MethodGet):
		// Resource changed via non-GET method, abandon monitoring.
		log.Info(
			"Abandoning provisioning state monitor, resource changed",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
		)

		return analyzer.Finished(), nil

	case !interaction.WasSuccessful(i):
		// Unexpected status code, abandon monitoring.
		log.Info(
			"Abandoning provisioning state monitor due to unexpected status",
			"url", m.baseURL.String(),
			"statusCode", statusCode,
		)

		return analyzer.Finished(), nil

	default:
		// GET with 2xx status - check provisioningState
		return m.checkProvisioningState(log, i)
	}
}

// checkProvisioningState examines the response body for provisioningState.
func (m *MonitorProvisioningState) checkProvisioningState(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Parse the response body to extract provisioningState
	var response azureResourceResponse

	err := json.Unmarshal(i.Response().Body(), &response)
	if err != nil {
		// Not a valid Azure resource response; abandon monitoring
		// This is not an error - just a condition this monitor isn't prepared to handle
		log.Info(
			"Abandoning provisioning state monitor, invalid JSON response",
			"url", m.baseURL.String(),
		)

		//nolint:nilerr // Invalid JSON is not an error, just a condition we can't handle
		return analyzer.Finished(), nil
	}

	currentState := response.Properties.ProvisioningState
	if currentState == "" {
		// No provisioningState field; abandon monitoring
		log.Info(
			"Abandoning provisioning state monitor, missing provisioningState",
			"url", m.baseURL.String(),
		)

		return analyzer.Finished(), nil
	}

	// Check if current state matches our target state (case-insensitive)
	if strings.EqualFold(currentState, m.targetState) {
		// Accumulate this interaction and continue monitoring
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil
	}

	// State has transitioned to a non-target state
	return m.stateTransitioned(log)
}

// stateTransitioned handles the case where provisioningState has moved to a final state.
func (m *MonitorProvisioningState) stateTransitioned(
	log logr.Logger,
) (analyzer.Result, error) {
	if len(m.interactions) < 2 {
		// No intermediate interactions to exclude.
		log.Info(
			"Short provisioning state sequence, nothing to exclude",
			"url", m.baseURL.String(),
		)

		return analyzer.Finished(), nil
	}

	log.Info(
		"Provisioning state sequence finished, excluding intermediate GETs",
		"url", m.baseURL.String(),
		"removed", len(m.interactions)-2,
	)

	excluded := m.interactions[1 : len(m.interactions)-1]

	return analyzer.FinishedWithExclusions(excluded...), nil
}
