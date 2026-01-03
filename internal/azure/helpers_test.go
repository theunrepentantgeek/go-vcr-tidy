package azure

import (
	"fmt"
	"net/url"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

/*
 * Helper functions for testing
 */

// runAnalyzer runs the analyzer with the provided interactions and returns the final result.
// It fails the test if any errors occur or if the analyzer finishes prematurely.
// t is the current test (don't pass a parent).
// a is the analyzer to run.
// interactions are the interactions to feed to the analyzer.
func runAnalyzer(
	t *testing.T,
	log logr.Logger,
	a analyzer.Interface,
	interactions ...interaction.Interface,
) analyzer.Result {
	t.Helper()
	g := NewWithT(t)

	var (
		result analyzer.Result
		err    error
	)

	limit := len(interactions) - 1
	for index, inter := range interactions {
		result, err = a.Analyze(log, inter)
		g.Expect(err).ToNot(HaveOccurred())

		if index < limit {
			g.Expect(result.Finished).To(BeFalse(), "Analyzer finished prematurely")
		}
	}

	return result
}

// mustParseURL parses a raw URL string and fails the test on error.
func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	return *parsed
}

// newTestLogger creates a test logger for the given test.
func newTestLogger(t *testing.T) logr.Logger {
	t.Helper()

	return testr.NewWithOptions(t, testr.Options{Verbosity: 1})
}

// createAzureResourceInteraction creates a fake interaction with Azure resource JSON.
// The response body will contain the specified provisioningState.
func createAzureResourceInteraction(
	fullURL url.URL,
	method string,
	statusCode int,
	provisioningState string,
) *fake.TestInteraction {
	jsonBody := fmt.Sprintf(
		`{"properties": {"provisioningState": "%s"}}`,
		provisioningState,
	)

	return createInteractionWithJSON(fullURL, method, statusCode, jsonBody)
}

// createInteractionWithJSON creates a fake interaction with the specified JSON body.
func createInteractionWithJSON(
	fullURL url.URL,
	method string,
	statusCode int,
	jsonBody string,
) *fake.TestInteraction {
	i := fake.Interaction(fullURL, method, statusCode)
	i.SetResponseBody(jsonBody)

	return i
}
