package azure

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
)

func TestDetectResourceDeletion_SuccessfulDELETE_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	deleteInteraction := createAzureResourceInteraction(baseURL, "DELETE", 200, "Deleting")
	result, err := detector.Analyze(log, deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorProvisioningState{}))
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestDetectResourceDeletion_Various2xxStatusCodes_SpawnsMonitor(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		"200 OK":         {statusCode: 200},
		"202 Accepted":   {statusCode: 202},
		"204 No Content": {statusCode: 204},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://management.azure.com/resource")
			detector := NewDetectResourceDeletion()
			log := newTestLogger(t)

			deleteInteraction := createAzureResourceInteraction(baseURL, "DELETE", c.statusCode, "Deleting")
			result, err := detector.Analyze(log, deleteInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result.Spawn).To(HaveLen(1))
		})
	}
}

func TestDetectResourceDeletion_FailedDELETE_DoesNotSpawn(t *testing.T) {
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
			detector := NewDetectResourceDeletion()
			log := newTestLogger(t)

			deleteInteraction := createAzureResourceInteraction(baseURL, "DELETE", c.statusCode, "Deleting")
			result, err := detector.Analyze(log, deleteInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectResourceDeletion_OtherMethods_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		"GET":   {method: "GET"},
		"POST":  {method: "POST"},
		"PUT":   {method: "PUT"},
		"PATCH": {method: "PATCH"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://management.azure.com/resource")
			detector := NewDetectResourceDeletion()
			log := newTestLogger(t)

			interaction := createAzureResourceInteraction(baseURL, c.method, 200, "Deleting")
			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectResourceDeletion_InvalidJSON_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	deleteInteraction := fake.Interaction(baseURL, "DELETE", 200)

	result, err := detector.Analyze(log, deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectResourceDeletion_MissingProvisioningState_DoesNotSpawn(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	deleteInteraction := createInteractionWithJSON(baseURL, "DELETE", 200, `{"properties": {}}`)

	result, err := detector.Analyze(log, deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

func TestDetectResourceDeletion_NeverFinishes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	// Process various interactions
	interactions := []*fake.TestInteraction{
		createAzureResourceInteraction(baseURL, "DELETE", 200, "Deleting"),
		createAzureResourceInteraction(baseURL, "DELETE", 202, "Deleting"),
		createAzureResourceInteraction(baseURL, "GET", 200, "Succeeded"),
	}

	for _, inter := range interactions {
		result, err := detector.Analyze(log, inter)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse(), "Detector should never finish")
	}
}

func TestDetectResourceDeletion_SpawnedMonitorHasCorrectState(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	deleteInteraction := createAzureResourceInteraction(baseURL, "DELETE", 200, "Deleting")
	result, err := detector.Analyze(log, deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))

	// Verify the spawned monitor is configured with correct state
	if m, ok := result.Spawn[0].(*MonitorProvisioningState); ok {
		g.Expect(m.targetStates).To(ContainElement("Deleting"))
		g.Expect(m.baseURL).To(Equal(baseURL))
	}
}

func TestDetectResourceDeletion_MultipleDELETEs_SpawnsMultipleMonitors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	url1 := mustParseURL("https://management.azure.com/resource/123")
	url2 := mustParseURL("https://management.azure.com/resource/456")

	delete1 := createAzureResourceInteraction(url1, "DELETE", 200, "Deleting")
	delete2 := createAzureResourceInteraction(url2, "DELETE", 202, "Deleting")

	result1, err := detector.Analyze(log, delete1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Spawn).To(HaveLen(1))

	result2, err := detector.Analyze(log, delete2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Spawn).To(HaveLen(1))

	// Each spawned monitor should be independent
	g.Expect(result1.Spawn[0]).ToNot(Equal(result2.Spawn[0]))
}

func TestDetectResourceDeletion_EmptyResult_WhenNoAction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://management.azure.com/resource")
	detector := NewDetectResourceDeletion()
	log := newTestLogger(t)

	getInteraction := createAzureResourceInteraction(baseURL, "GET", 200, "Succeeded")
	result, err := detector.Analyze(log, getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
