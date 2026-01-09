package generic

import (
	"net/http"
	"testing"

	"github.com/neilotoole/slogt"
	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestMonitorDeletion_SingleGETReturning404_MarksFinished(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Single GET returning 404 should finish immediately
	i := fake.Interaction(baseURL, http.MethodGet, 404)

	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty(), "Single 404 should not exclude any interactions")
}

func TestMonitorDeletion_TwoGETsThenConfirmation_NothingIsRemoved(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Two successful GETs followed by 404
	get1 := fake.Interaction(baseURL, http.MethodGet, 200)
	get2 := fake.Interaction(baseURL, http.MethodGet, 200)
	get404 := fake.Interaction(baseURL, http.MethodGet, 404)

	result := runAnalyzer(t, log, monitor, get1, get2, get404)
	g.Expect(result.Finished).To(BeTrue())

	// With only 2 accumulated interactions, none should be excluded
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorDeletion_ThreeGETsThenConfirmation_MiddleIsRemoved(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Three successful GETs followed by 404
	get1 := fake.Interaction(baseURL, http.MethodGet, 200)
	get2 := fake.Interaction(baseURL, http.MethodGet, 200)
	get3 := fake.Interaction(baseURL, http.MethodGet, 200)
	get404 := fake.Interaction(baseURL, http.MethodGet, 404)

	result := runAnalyzer(t, log, monitor, get1, get2, get3, get404)
	g.Expect(result.Finished).To(BeTrue())

	// First and last accumulated should remain, middle should be excluded
	g.Expect(result.Excluded).To(HaveLen(1))
	g.Expect(result.Excluded).To(ContainElement(get2))
}

func TestMonitorDeletion_MultipleMiddleGETs_AllMiddleAreRemoved(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Many successful GETs followed by 404
	interactions := make([]interaction.Interface, 0, 10)

	for range 9 {
		get := fake.Interaction(baseURL, http.MethodGet, 200)
		interactions = append(interactions, get)
	}

	get404 := fake.Interaction(baseURL, http.MethodGet, 404)
	interactions = append(interactions, get404)

	result := runAnalyzer(t, log, monitor, interactions...)
	g.Expect(result.Finished).To(BeTrue())

	// Verify exclusion pattern: all middle interactions should be excluded
	g.Expect(result.Excluded).To(HaveLen(7))

	for i := 1; i < 8; i++ {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]), "Middle GET %d should be excluded", i)
	}
}

func TestMonitorDeletion_DifferentURL_Ignored(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://api.example.com/resource/123")
	differentURL := must.ParseURL(t, "https://api.example.com/resource/456")
	monitor := NewMonitorDeletion(monitoredURL)
	log := slogt.New(t)

	i := fake.Interaction(differentURL, http.MethodGet, 200)

	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestMonitorDeletion_AbandonsMonitoring(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method     string
		statusCode int
	}{
		"POST":   {method: http.MethodPost, statusCode: 201},
		"PUT":    {method: http.MethodPut, statusCode: 200},
		"GET500": {method: http.MethodGet, statusCode: 500},
		"GET301": {method: http.MethodGet, statusCode: 301},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			monitor := NewMonitorDeletion(baseURL)
			log := slogt.New(t)

			// Start with some successful GETs
			// Then a request that should abandon monitoring
			get1 := fake.Interaction(baseURL, http.MethodGet, 200)
			abandoningRequest := fake.Interaction(baseURL, c.method, c.statusCode)

			result := runAnalyzer(t, log, monitor, get1, abandoningRequest)
			g.Expect(result.Finished).To(BeTrue())
			g.Expect(result.Excluded).To(BeEmpty())
		})
	}
}

func TestMonitorDeletion_Various2xxStatusCodes_Accumulated(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Test various 2xx status codes
	statusCodes := []int{200, 201, 202, 204, 206}
	interactions := make([]interaction.Interface, 0, len(statusCodes))

	for _, code := range statusCodes {
		get := fake.Interaction(baseURL, http.MethodGet, code)
		interactions = append(interactions, get)
	}

	get404 := fake.Interaction(baseURL, http.MethodGet, 404)
	interactions = append(interactions, get404)

	result := runAnalyzer(t, log, monitor, interactions...)

	// First and last accumulated should remain, middle should be excluded
	g.Expect(result.Excluded).To(HaveLen(3))

	for i := 1; i < len(statusCodes)-1; i++ {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]))
	}
}

func TestMonitorDeletion_URLWithQueryParameters_MonitorsBaseURL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	urlWithParams := must.ParseURL(t, "https://api.example.com/resource/123?param=value")
	monitor := NewMonitorDeletion(baseURL)
	log := slogt.New(t)

	// Interaction with query parameters should match base URL
	i := fake.Interaction(urlWithParams, http.MethodGet, 200)
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
}

func TestMonitorDeletion_EmptyResult_WhenIgnoringInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitoredURL := must.ParseURL(t, "https://api.example.com/resource/123")
	differentURL := must.ParseURL(t, "https://api.example.com/other")
	monitor := NewMonitorDeletion(monitoredURL)
	log := slogt.New(t)

	i := fake.Interaction(differentURL, http.MethodGet, 200)
	result, err := monitor.Analyze(log, i)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
