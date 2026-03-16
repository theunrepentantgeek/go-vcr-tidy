package generic

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

func TestMonitorDeferredCreation_SingleGETThenSuccess_MarksFinished(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	get200 := fake.Interaction(baseURL, http.MethodGet, 200)
	result, err := monitor.Analyze(log, get200)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty(), "Single 404 should not exclude any interactions")
}

func TestMonitorDeferredCreation_Two404sThenSuccess_NothingIsRemoved(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	second404 := fake.Interaction(baseURL, http.MethodGet, 404)
	get200 := fake.Interaction(baseURL, http.MethodGet, 200)

	result := runAnalyzer(t, log, monitor, second404, get200)
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorDeferredCreation_Three404sThenSuccess_MiddleIsRemoved(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	second404 := fake.Interaction(baseURL, http.MethodGet, 404)
	third404 := fake.Interaction(baseURL, http.MethodGet, 404)
	get200 := fake.Interaction(baseURL, http.MethodGet, 200)

	result := runAnalyzer(t, log, monitor, second404, third404, get200)
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(HaveLen(1))
	g.Expect(result.Excluded).To(ContainElement(second404))
}

func TestMonitorDeferredCreation_MultipleMiddle404s_AllMiddleAreRemoved(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	interactions := make([]interaction.Interface, 0, 10)

	for range 8 {
		get := fake.Interaction(baseURL, http.MethodGet, 404)
		interactions = append(interactions, get)
	}

	get200 := fake.Interaction(baseURL, http.MethodGet, 200)
	interactions = append(interactions, get200)

	result := runAnalyzer(t, log, monitor, interactions...)
	g.Expect(result.Finished).To(BeTrue())

	// 9 total 404s (1 seed + 8 more), exclude middle 7
	g.Expect(result.Excluded).To(HaveLen(7))

	for i := range 7 {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]), "Middle 404 %d should be excluded", i)
	}
}

func TestMonitorDeferredCreation_DifferentURL_Ignored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://api.example.com/resource/123")
	differentURL := must.ParseURL(t, "https://api.example.com/resource/456")
	first404 := fake.Interaction(monitoredURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	i := fake.Interaction(differentURL, http.MethodGet, 200)
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorDeferredCreation_AbandonsMonitoring(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method     string
		statusCode int
	}{
		"POST":   {method: http.MethodPost, statusCode: 201},
		"PUT":    {method: http.MethodPut, statusCode: 200},
		"DELETE": {method: http.MethodDelete, statusCode: 204},
		"PATCH":  {method: http.MethodPatch, statusCode: 200},
		"GET500": {method: http.MethodGet, statusCode: 500},
		"GET301": {method: http.MethodGet, statusCode: 301},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			first404 := fake.Interaction(baseURL, http.MethodGet, 404)
			monitor := NewMonitorDeferredCreation(first404)
			log := slogt.New(t)

			abandoningRequest := fake.Interaction(baseURL, c.method, c.statusCode)
			result, err := monitor.Analyze(log, abandoningRequest)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result.Finished).To(BeTrue())
			g.Expect(result.Excluded).To(BeEmpty())
		})
	}
}

func TestMonitorDeferredCreation_Various2xxStatusCodes_ConfirmCreation(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	// Add enough 404s to test exclusion
	second404 := fake.Interaction(baseURL, http.MethodGet, 404)
	third404 := fake.Interaction(baseURL, http.MethodGet, 404)

	result1, err := monitor.Analyze(log, second404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Finished).To(BeFalse())

	result2, err := monitor.Analyze(log, third404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Finished).To(BeFalse())

	// 201 should also confirm creation
	get201 := fake.Interaction(baseURL, http.MethodGet, 201)
	result3, err := monitor.Analyze(log, get201)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result3.Finished).To(BeTrue())
	g.Expect(result3.Excluded).To(HaveLen(1))
	g.Expect(result3.Excluded).To(ContainElement(second404))
}

func TestMonitorDeferredCreation_URLWithQueryParameters_MonitorsBaseURL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	first404 := fake.Interaction(baseURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	urlWithParams := must.ParseURL(t, "https://api.example.com/resource/123?param=value")
	i := fake.Interaction(urlWithParams, http.MethodGet, 404)
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
}

func TestMonitorDeferredCreation_EmptyResult_WhenIgnoringInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://api.example.com/resource/123")
	differentURL := must.ParseURL(t, "https://api.example.com/other")
	first404 := fake.Interaction(monitoredURL, http.MethodGet, 404)
	monitor := NewMonitorDeferredCreation(first404)
	log := slogt.New(t)

	i := fake.Interaction(differentURL, http.MethodGet, 200)
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
