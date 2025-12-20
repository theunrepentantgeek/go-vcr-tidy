package azure

import (
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectAzureLongRunningOperation is an analyzer for detecting Azure long-running operations.
// It watches for successful PUT, POST or DELETE requests where the response includes a `Azure-Asyncoperation` header.
type DetectAzureLongRunningOperation struct{}

var _ analyzer.Interface = &DetectAzureLongRunningOperation{}

// NewDetectAzureLongRunningOperation creates a new DetectAzureLongRunningOperation analyzer.
func NewDetectAzureLongRunningOperation() *DetectAzureLongRunningOperation {
	return &DetectAzureLongRunningOperation{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (*DetectAzureLongRunningOperation) Analyze(
	log logr.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	// Check if the interaction is a PUT, POST, or DELETE
	method := i.Request().Method()
	if method != http.MethodPut &&
		method != http.MethodPost &&
		method != http.MethodDelete {
		return analyzer.Result{}, nil
	}

	// Check if the response is successful
	statusCode := i.Response().StatusCode()
	if statusCode < 200 || statusCode >= 300 {
		return analyzer.Result{}, nil
	}

	// Check for the Azure-Asyncoperation header
	asyncHeader, ok := i.Response().Header("Azure-Asyncoperation")
	if !ok || asyncHeader == "" {
		return analyzer.Result{}, nil
	}

	operationURL, err := url.Parse(asyncHeader)
	if err != nil {
		return analyzer.Result{}, eris.Wrapf(err, "parsing Azure-Asyncoperation URL: %s", asyncHeader)
	}

	log.Info(
		"Found Azure long running operation",
		"url", operationURL)

	monitor := NewMonitorAzureLongRunningOperation(*operationURL)

	return analyzer.Spawn(monitor), nil
}
