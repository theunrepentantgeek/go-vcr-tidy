package azure

import (
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectResourceDeletion is an analyzer for detecting Azure resource deletion.
// It watches for successful DELETE requests to Azure resources and spawns a MonitorProvisioningState
// analyzer to track subsequent GET requests monitoring for "Deleting" provisioning state.
type DetectResourceDeletion struct{}

var _ analyzer.Interface = &DetectResourceDeletion{}

// NewDetectResourceDeletion creates a new DetectResourceDeletion analyzer.
func NewDetectResourceDeletion() *DetectResourceDeletion {
	return &DetectResourceDeletion{}
}

// Analyze processes another interaction in the sequence.
func (*DetectResourceDeletion) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	statusCode := i.Response().StatusCode()

	// Check if it's a DELETE with successful status
	if !interaction.HasMethod(i, http.MethodDelete) || statusCode < 200 || statusCode >= 300 {
		return analyzer.Result{}, nil
	}

	// Check if response contains provisioningState indicating deletion
	var response azureResourceResponse

	err := json.Unmarshal(i.Response().Body(), &response)
	if err != nil {
		// Not a valid Azure resource response; ignore
		//nolint:nilerr // Just ignore invalid responses
		return analyzer.Result{}, nil
	}

	// Only spawn monitor if we have a provisioningState
	if response.Properties.ProvisioningState == "" {
		return analyzer.Result{}, nil
	}

	// Start monitoring for Deleting state
	reqURL := i.Request().BaseURL()
	log.Info(
		"Found resource deletion to monitor",
		"url", reqURL.String(),
		"provisioningState", response.Properties.ProvisioningState,
	)

	monitor := NewMonitorProvisioningState(reqURL, "Deleting")

	return analyzer.Spawn(monitor), nil
}
