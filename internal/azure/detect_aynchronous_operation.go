package azure

import (
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// DetectAzureAsynchronousOperation is an analyzer for detecting Azure asynchronous operations.
// It watches for successful PUT, POST or DELETE requests where the response includes a `Azure-Asyncoperation` header.
type DetectAzureAsynchronousOperation struct{}

const azureLocationHeader = "Location"

var _ analyzer.Interface = &DetectAzureAsynchronousOperation{}

// NewDetectAzureAsynchronousOperation creates a new DetectAzureAsynchronousOperation analyzer.
func NewDetectAzureAsynchronousOperation() *DetectAzureAsynchronousOperation {
	return &DetectAzureAsynchronousOperation{}
}

// Analyze processes another interaction in the sequence.
// interaction is the interaction to analyze.
func (*DetectAzureAsynchronousOperation) Analyze(
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

	// Check if the response is 202 Accepted
	statusCode := i.Response().StatusCode()
	if statusCode != http.StatusAccepted {
		return analyzer.Result{}, nil
	}

	// Check for the Location header
	locationHeader, ok := i.Response().Header(azureLocationHeader)
	if !ok || locationHeader == "" {
		return analyzer.Result{}, nil
	}

	operationURL, err := url.Parse(locationHeader)
	if err != nil {
		return analyzer.Result{}, eris.Wrapf(err, "parsing Location URL: %s", locationHeader)
	}

	log.Info(
		"Found Azure asynchronous operation",
		"url", urltool.BaseURL(operationURL).String())

	monitor := NewMonitorAzureAsynchronousOperation(operationURL)

	return analyzer.Spawn(monitor), nil
}
