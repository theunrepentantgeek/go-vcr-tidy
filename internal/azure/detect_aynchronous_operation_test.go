package azure

import (
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
		"PUT":    {method: "PUT"},
		"POST":   {method: "POST"},
		"DELETE": {method: "DELETE"},
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
		"GET":   {method: "GET"},
		"PATCH": {method: "PATCH"},
		"HEAD":  {method: "HEAD"},
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
		"200 OK":           {statusCode: 200},
		"201 Created":      {statusCode: 201},
		"204 No Content":   {statusCode: 204},
		"400 Bad Request":  {statusCode: 400},
		"404 Not Found":    {statusCode: 404},
		"500 Server Error": {statusCode: 500},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			locationURL := testAsyncOperationURL
			detector := NewDetectAzureAsynchronousOperation()
			log := slogt.New(t)

			interaction := fake.Interaction(baseURL, "PUT", c.statusCode)
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

	interaction := fake.Interaction(baseURL, "PUT", 202)

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

	interaction := fake.Interaction(baseURL, "PUT", 202)
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

	interaction := fake.Interaction(baseURL, "PUT", 202)
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

	interaction := fake.Interaction(baseURL, "PUT", 202)
	interaction.SetResponseHeader(azureLocationHeader, locationURL)

	result, err := detector.Analyze(log, interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))

	monitor, ok := result.Spawn[0].(*MonitorAzureAsynchronousOperation)
	g.Expect(ok).To(BeTrue(), "Expected spawned monitor to be *MonitorAzureAsynchronousOperation")
	g.Expect(monitor.operationURL.String()).To(Equal(locationURL))
}
