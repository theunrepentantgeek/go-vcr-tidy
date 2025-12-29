package azure

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorProvisioningState is an analyzer for monitoring Azure resource provisioning states.
// It watches for GET requests to a specific URL and tracks interactions where the provisioningState
// matches one of the target states (case-insensitive). When the provisioningState transitions to a
// different value, the monitor finishes and excludes all but the first and last accumulated interactions.
type MonitorProvisioningState struct {
	baseURL      url.URL                 // Base URL of the resource to monitor
	targetStates []string                // States to monitor (e.g., "Creating", "Updating")
	interactions []interaction.Interface // Accumulated interactions with matching provisioningState
}

var _ analyzer.Interface = (*MonitorProvisioningState)(nil)

// NewMonitorProvisioningState creates a new MonitorProvisioningState analyzer.
// baseURL is the base URL of the resource to monitor.
// targetStates are the provisioningState values to watch for (case-insensitive).
func NewMonitorProvisioningState(
	baseURL url.URL,
	targetStates []string,
) *MonitorProvisioningState {
	return &MonitorProvisioningState{
		baseURL:      baseURL,
		targetStates: targetStates,
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorProvisioningState) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()
	method := i.Request().Method()
	statusCode := i.Response().StatusCode()

	switch {
	case !urltool.SameBaseURL(reqURL, m.baseURL):
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil

	case method != "GET":
		// Resource changed via non-GET method, abandon monitoring.
		log.Info(
			"Abandoning provisioning state monitor, resource changed",
			"url", m.baseURL.String(),
			"method", method,
		)

		return analyzer.Finished(), nil

	case statusCode < 200 || statusCode >= 300:
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
		// Not a valid Azure resource response; ignore
		//nolint:nilerr // Just ignore invalid responses
		return analyzer.Result{}, nil
	}

	currentState := response.Properties.ProvisioningState
	if currentState == "" {
		// No provisioningState field; ignore
		return analyzer.Result{}, nil
	}

	// Check if current state matches any of our target states (case-insensitive)
	if m.isTargetState(currentState) {
		// Accumulate this interaction and continue monitoring
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil
	}

	// State has transitioned to a non-target state
	return m.stateTransitioned(log)
}

// isTargetState checks if the given state matches any target state (case-insensitive).
func (m *MonitorProvisioningState) isTargetState(state string) bool {
	for _, target := range m.targetStates {
		if strings.EqualFold(state, target) {
			return true
		}
	}

	return false
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
