package azure

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
)

func TestDetectResourceModification_SuccessfulPUT_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	putInteraction := createAzureResourceInteraction(baseURL, "PUT", 200, "Creating")
	result, err := detector.Analyze(log, putInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorProvisioningState{}))
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestDetectResourceModification_SuccessfulPATCH_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	patchInteraction := createAzureResourceInteraction(baseURL, "PATCH", 200, "Updating")
	result, err := detector.Analyze(log, patchInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorProvisioningState{}))
}

func TestDetectResourceModification_Various2xxStatusCodes_SpawnsMonitor(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		"200 OK":         {statusCode: 200},
		"201 Created":    {statusCode: 201},
		"202 Accepted":   {statusCode: 202},
		"204 No Content": {statusCode: 204},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://management.azure.com/resource")
			detector := NewDetectResourceModification()
			log := newTestLogger(t)

			putInteraction := createAzureResourceInteraction(baseURL, "PUT", c.statusCode, "Creating")
			result, err := detector.Analyze(log, putInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result.Spawn).To(HaveLen(1))
		})
	}
}

func TestDetectResourceModification_FailedRequest_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		"400 Bad Request":  {statusCode: 400},
		"404 Not Found":    {statusCode: 404},
		"500 Server Error": {statusCode: 500},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://management.azure.com/resource")
			detector := NewDetectResourceModification()
			log := newTestLogger(t)

			putInteraction := createAzureResourceInteraction(baseURL, "PUT", c.statusCode, "Creating")
			result, err := detector.Analyze(log, putInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectResourceModification_OtherMethods_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		"GET":    {method: "GET"},
		"POST":   {method: "POST"},
		"DELETE": {method: "DELETE"},
		"HEAD":   {method: "HEAD"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://management.azure.com/resource")
			detector := NewDetectResourceModification()
			log := newTestLogger(t)

			interaction := createAzureResourceInteraction(baseURL, c.method, 200, "Creating")
			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectResourceModification_InvalidJSON_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	putInteraction := fake.Interaction(baseURL, "PUT", 200)

	result, err := detector.Analyze(log, putInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectResourceModification_MissingProvisioningState_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	putInteraction := createInteractionWithJSON(baseURL, "PUT", 200, `{"properties": {}}`)

	result, err := detector.Analyze(log, putInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectResourceModification_NeverFinishes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	// Process various interactions
	interactions := []*fake.TestInteraction{
		createAzureResourceInteraction(baseURL, "GET", 200, "Succeeded"),
		createAzureResourceInteraction(baseURL, "PUT", 200, "Creating"),
		createAzureResourceInteraction(baseURL, "PATCH", 200, "Updating"),
		createAzureResourceInteraction(baseURL, "DELETE", 200, "Deleting"),
	}

	for _, inter := range interactions {
		result, err := detector.Analyze(log, inter)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse(), "Detector should never finish")
	}
}

func TestDetectResourceModification_SpawnedMonitorHasCorrectStates(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	putInteraction := createAzureResourceInteraction(baseURL, "PUT", 200, "Creating")
	result, err := detector.Analyze(log, putInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))

	// Verify the spawned monitor is configured with correct states
	if m, ok := result.Spawn[0].(*MonitorProvisioningState); ok {
		g.Expect(m.targetStates).To(ContainElements("Creating", "Updating"))
		g.Expect(m.baseURL).To(Equal(baseURL))
	}
}

func TestDetectResourceModification_MultipleRequests_SpawnsMultipleMonitors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	url1 := mustParseURL("https://management.azure.com/resource/123")
	url2 := mustParseURL("https://management.azure.com/resource/456")

	put1 := createAzureResourceInteraction(url1, "PUT", 200, "Creating")
	put2 := createAzureResourceInteraction(url2, "PUT", 201, "Creating")

	result1, err := detector.Analyze(log, put1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Spawn).To(HaveLen(1))

	result2, err := detector.Analyze(log, put2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Spawn).To(HaveLen(1))

	// Each spawned monitor should be independent
	g.Expect(result1.Spawn[0]).ToNot(Equal(result2.Spawn[0]))
}

func TestDetectResourceModification_EmptyResult_WhenNoAction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceModification()
	log := newTestLogger(t)

	getInteraction := createAzureResourceInteraction(baseURL, "GET", 200, "Succeeded")
	result, err := detector.Analyze(log, getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
