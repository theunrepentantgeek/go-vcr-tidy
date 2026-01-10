package azure

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// DetectAzureLongRunningOperation is an analyzer for detecting Azure long-running operations.
// It watches for successful PUT, POST or DELETE requests where the response includes a `Azure-Asyncoperation` header.
type DetectAzureLongRunningOperation struct{}

const azureLROHeader = "Azure-Asyncoperation"

var _ analyzer.Interface = &DetectAzureLongRunningOperation{}

// NewDetectAzureLongRunningOperation creates a new DetectAzureLongRunningOperation analyzer.
func NewDetectAzureLongRunningOperation() *DetectAzureLongRunningOperation {
	return &DetectAzureLongRunningOperation{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (*DetectAzureLongRunningOperation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Check if the interaction is a PUT, POST, or DELETE
	if !interaction.HasAnyMethod(i, http.MethodPut, http.MethodPost, http.MethodDelete) {
		return analyzer.Result{}, nil
	}

	// Check if the response is successful
	if !interaction.WasSuccessful(i) {
		return analyzer.Result{}, nil
	}

	// Check for the Azure-Asyncoperation header
	asyncHeader, ok := i.Response().Header(azureLROHeader)
	if !ok || asyncHeader == "" {
		return analyzer.Result{}, nil
	}

	operationURL, err := url.Parse(asyncHeader)
	if err != nil {
		return analyzer.Result{}, eris.Wrapf(err, "parsing Azure-Asyncoperation URL: %s", asyncHeader)
	}

	log.Debug(
		"Found Azure long running operation",
		"url", urltool.BaseURL(operationURL).String())

	monitor := NewMonitorAzureLongRunningOperation(operationURL)

	return analyzer.Spawn(monitor), nil
}
