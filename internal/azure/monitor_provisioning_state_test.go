package azure

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/neilotoole/slogt"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestMonitorProvisioningState_SingleState_AccumulatesAndExcludes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Create interactions with provisioningState in response
	get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
	get2 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
	get3 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Succeeded")

	result := runAnalyzer(t, log, monitor, get1, get2, get3, getFinal)

	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(HaveLen(1))
	g.Expect(result.Excluded).To(ContainElement(get2))
}

func TestMonitorProvisioningState_OnlyMatchesSpecificState(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Monitor should only accumulate "Creating" states, not "Updating"
	get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
	get2 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Updating")

	result1, err := monitor.Analyze(log, get1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Finished).To(BeFalse())
	g.Expect(monitor.interactions).To(HaveLen(1))

	// When state changes to "Updating", monitor should finish
	result2, err := monitor.Analyze(log, get2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Finished).To(BeTrue())
	// Only one "Creating" interaction, so nothing to exclude
	g.Expect(result2.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_CaseInsensitive_MatchesState(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Various case combinations
	get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "creating")
	get2 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "CREATING")
	get3 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "CrEaTiNg")
	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Succeeded")

	result := runAnalyzer(t, log, monitor, get1, get2, get3, getFinal)

	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(HaveLen(1))
	g.Expect(result.Excluded).To(ContainElement(get2))
}

func TestMonitorProvisioningState_ShortSequence_NothingExcluded(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Only one Creating state before transition
	get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Succeeded")

	result := runAnalyzer(t, log, monitor, get1, getFinal)

	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_ImmediateTransition_NothingExcluded(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Immediate transition without any Creating states
	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Succeeded")

	result, err := monitor.Analyze(log, getFinal)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_DifferentURL_Ignored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://management.azure.com/resource/123")
	differentURL := must.ParseURL(t, "https://management.azure.com/resource/456")
	monitor := NewMonitorProvisioningState(monitoredURL, "Creating")
	log := slogt.New(t)

	i := createAzureResourceInteraction(differentURL, http.MethodGet, 200, "Creating")

	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_NonGETMethod_AbandonsMonitoring(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		"POST":   {method: http.MethodPost},
		"PUT":    {method: http.MethodPut},
		"PATCH":  {method: http.MethodPatch},
		"DELETE": {method: http.MethodDelete},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			monitor := NewMonitorProvisioningState(baseURL, "Creating")
			log := slogt.New(t)

			// Accumulate some interactions first
			get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
			modifyRequest := createAzureResourceInteraction(baseURL, c.method, 200, "Creating")

			result := runAnalyzer(t, log, monitor, get1, modifyRequest)

			g.Expect(result.Finished).To(BeTrue())
			g.Expect(result.Excluded).To(BeEmpty())
		})
	}
}

func TestMonitorProvisioningState_UnexpectedStatusCode_AbandonsMonitoring(t *testing.T) {
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

			baseURL := must.ParseURL(t, "https://management.azure.com/resource")
			monitor := NewMonitorProvisioningState(baseURL, "Creating")
			log := slogt.New(t)

			get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
			getError := fake.Interaction(baseURL, http.MethodGet, c.statusCode)

			result := runAnalyzer(t, log, monitor, get1, getError)

			g.Expect(result.Finished).To(BeTrue())
			g.Expect(result.Excluded).To(BeEmpty())
		})
	}
}

func TestMonitorProvisioningState_InvalidJSON_AbandonMonitoring(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Create interaction with invalid JSON
	getInvalid := fake.Interaction(baseURL, http.MethodGet, 200)

	result, err := monitor.Analyze(log, getInvalid)

	g.Expect(err).ToNot(HaveOccurred(), "Invalid JSON should not be treated as an error")
	g.Expect(result.Finished).To(BeTrue(), "Should abandon monitoring on invalid JSON")
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_MissingProvisioningState_AbandonMonitoring(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Create interaction with valid JSON but no provisioningState
	getNoState := createInteractionWithJSON(baseURL, http.MethodGet, 200, `{"properties": {}}`)

	result, err := monitor.Analyze(log, getNoState)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue(), "Should abandon monitoring when provisioningState is missing")
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorProvisioningState_URLWithQueryParameters_MonitorsBaseURL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	urlWithParams := must.ParseURL(t, "https://management.azure.com/resource?api-version=2021-01-01")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Interaction with query parameters should match base URL
	get1 := createAzureResourceInteraction(urlWithParams, http.MethodGet, 200, "Creating")

	result, err := monitor.Analyze(log, get1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	// Interaction should be accumulated
	g.Expect(monitor.interactions).To(HaveLen(1))
}

func TestMonitorProvisioningState_ManyMiddleInteractions_AllExcluded(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Creating")
	log := slogt.New(t)

	// Create many interactions
	interactions := make([]interaction.Interface, 0, 10)

	for range 9 {
		get := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Creating")
		interactions = append(interactions, get)
	}

	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Succeeded")
	interactions = append(interactions, getFinal)

	result := runAnalyzer(t, log, monitor, interactions...)

	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(HaveLen(7))

	// Verify the middle interactions are excluded
	for i := 1; i < 8; i++ {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]))
	}
}

func TestMonitorProvisioningState_DeletingState_Works(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://management.azure.com/resource")
	monitor := NewMonitorProvisioningState(baseURL, "Deleting")
	log := slogt.New(t)

	get1 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Deleting")
	get2 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Deleting")
	get3 := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Deleting")
	getFinal := createAzureResourceInteraction(baseURL, http.MethodGet, 200, "Deleted")

	result := runAnalyzer(t, log, monitor, get1, get2, get3, getFinal)

	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(HaveLen(1))
	g.Expect(result.Excluded).To(ContainElement(get2))
}

func TestMonitorProvisioningState_EmptyResult_WhenIgnoringInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://management.azure.com/resource/123")
	differentURL := must.ParseURL(t, "https://management.azure.com/other")
	monitor := NewMonitorProvisioningState(monitoredURL, "Creating")
	log := slogt.New(t)

	i := createAzureResourceInteraction(differentURL, http.MethodGet, 200, "Creating")
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
