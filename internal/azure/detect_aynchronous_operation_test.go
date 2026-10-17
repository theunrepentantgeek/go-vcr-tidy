package azure

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/neilotoole/slogt"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

const testAsyncOperationURL = "https://management.azure.com/operations/12345"

func TestDetectAzureAsynchronousOperation_SuccessfulRequestsWith202AndLocationHeader_SpawnsMonitor(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		http.MethodPut:    {method: http.MethodPut},
		http.MethodPost:   {method: http.MethodPost},
		http.MethodDelete: {method: http.MethodDelete},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			locationURL := testAsyncOperationURL
			detector := NewDetectAzureAsynchronousOperation()
			log := slogt.New(t)

			interaction := fake.Interaction(baseURL, c.method, 202)
			interaction.SetResponseHeader(azureLocationHeader, locationURL)

			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result.Finished).To(BeFalse())
			g.Expect(result.Spawn).To(HaveLen(1))
			g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorAzureAsynchronousOperation{}))
			g.Expect(result.Excluded).To(BeEmpty())
		})
	}
}

func TestDetectAzureAsynchronousOperation_OtherMethods_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		http.MethodGet:   {method: http.MethodGet},
		http.MethodPatch: {method: http.MethodPatch},
		http.MethodHead:  {method: http.MethodHead},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			locationURL := testAsyncOperationURL
			detector := NewDetectAzureAsynchronousOperation()
			log := slogt.New(t)

			interaction := fake.Interaction(baseURL, c.method, 202)
			interaction.SetResponseHeader(azureLocationHeader, locationURL)

			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectAzureAsynchronousOperation_Non202StatusCode_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		testStatusOK:          {statusCode: 200},
		"201 Created":         {statusCode: 201},
		testStatusNoContent:   {statusCode: 204},
		testStatusBadRequest:  {statusCode: 400},
		testStatusNotFound:    {statusCode: 404},
		testStatusServerError: {statusCode: 500},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			locationURL := testAsyncOperationURL
			detector := NewDetectAzureAsynchronousOperation()
			log := slogt.New(t)

			interaction := fake.Interaction(baseURL, http.MethodPut, c.statusCode)
			interaction.SetResponseHeader(azureLocationHeader, locationURL)

			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectAzureAsynchronousOperation_MissingLocationHeader_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	detector := NewDetectAzureAsynchronousOperation()
	log := slogt.New(t)

	interaction := fake.Interaction(baseURL, http.MethodPut, 202)

	result, err := detector.Analyze(log, interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectAzureAsynchronousOperation_EmptyLocationHeader_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	detector := NewDetectAzureAsynchronousOperation()
	log := slogt.New(t)

	interaction := fake.Interaction(baseURL, http.MethodPut, 202)
	interaction.SetResponseHeader(azureLocationHeader, "")

	result, err := detector.Analyze(log, interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectAzureAsynchronousOperation_InvalidLocationURL_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	detector := NewDetectAzureAsynchronousOperation()
	log := slogt.New(t)

	interaction := fake.Interaction(baseURL, http.MethodPut, 202)
	interaction.SetResponseHeader(azureLocationHeader, "://invalid")

	result, err := detector.Analyze(log, interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectAzureAsynchronousOperation_ValidLocationURL_ParsesCorrectly(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	locationURL := "https://management.azure.com/subscriptions/abc123/providers/" +
		"Microsoft.Network/locations/eastus/operations/xyz789"
	detector := NewDetectAzureAsynchronousOperation()
	log := slogt.New(t)

	interaction := fake.Interaction(baseURL, http.MethodPut, 202)
	interaction.SetResponseHeader(azureLocationHeader, locationURL)

	result, err := detector.Analyze(log, interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))

	monitor, ok := result.Spawn[0].(*MonitorAzureAsynchronousOperation)
	g.Expect(ok).To(BeTrue(), "Expected spawned monitor to be *MonitorAzureAsynchronousOperation")
	g.Expect(monitor.operationURL.String()).To(Equal(locationURL))
}
