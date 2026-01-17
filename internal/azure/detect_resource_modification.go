package azure

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectResourceModification is an analyzer for detecting Azure resource creation and updates.
// It watches for successful PUT or PATCH requests to Azure resources and spawns a MonitorProvisioningState
// analyzer to track subsequent GET requests monitoring for "Creating" or "Updating" provisioning states.
type DetectResourceModification struct{}

var _ analyzer.Interface = &DetectResourceModification{}

// NewDetectResourceModification creates a new DetectResourceModification analyzer.
func NewDetectResourceModification() *DetectResourceModification {
	return &DetectResourceModification{}
}

// Analyze processes another interaction in the sequence.
func (*DetectResourceModification) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Check if it's a PUT or PATCH with successful status
	if !interaction.HasAnyMethod(i, http.MethodPut, http.MethodPatch) || !interaction.WasSuccessful(i) {
		return analyzer.Result{}, nil
	}

	// Check if response contains provisioningState indicating a transient state
	var response ResourceResponse

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

	// Start monitoring for Creating/Updating states
	reqURL := i.Request().BaseURL()
	log.Debug(
		"Found resource modification to monitor",
		"url", reqURL.String(),
		"method", i.Request().Method(),
		"provisioningState", response.Properties.ProvisioningState,
	)

	monitorCreating := NewMonitorProvisioningState(reqURL, "Creating")
	monitorUpdating := NewMonitorProvisioningState(reqURL, "Updating")

	return analyzer.Spawn(monitorCreating, monitorUpdating), nil
}
