package azure

import (
	"net/http"
	"net/url"

	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"

)

// DetectAzureLongRunningOperation is an analyzer for detecting Azure long-running operations.
// It watches for successful PUT, POST or DELETE requests where the response includes a `Azure-Asyncoperation` header.
type DetectAzureLongRunningOperation struct {
}

var _ analyzer.Interface = &DetectAzureLongRunningOperation{}

// NewDetectAzureLongRunningOperation creates a new DetectAzureLongRunningOperation analyzer.
func NewDetectAzureLongRunningOperation() *DetectAzureLongRunningOperation {
	return &DetectAzureLongRunningOperation{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (d *DetectAzureLongRunningOperation) Analyze(
	_ logr.Logger,
	interaction interaction.Interface,
) (analyzer.Result, error) {
	// Check if the interaction is a PUT, POST, or DELETE
	method := interaction.Request().Method()
	if method != http.MethodPut &&
		method != http.MethodPost &&
		method != http.MethodDelete {
		return analyzer.Result{}, nil
	}

	// Check if the response is successful
	statusCode := interaction.Response().StatusCode()
	if statusCode < 200 || statusCode >= 300 {
		return analyzer.Result{}, nil
	}

	// Check for the Azure-Asyncoperation header
	asyncHeader, ok := interaction.Response().Header("Azure-Asyncoperation")
	if !ok || asyncHeader == "" {
		return analyzer.Result{}, nil
	}

	operationURL, err := url.Parse(asyncHeader)
	if err != nil {
		return analyzer.Result{}, err
	}

	monitor := NewMonitorAzureLongRunningOperation(*operationURL)
	return analyzer.Spawn(monitor), nil
}
