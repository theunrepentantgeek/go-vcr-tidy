package generic_test

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/generic"
)

func TestMonitorDeletion_SingleGETReturning404_MarksFinished(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Single GET returning 404 should finish immediately
	interaction := fake.NewInteraction(baseURL, "GET", 404)

	result, err := monitor.Analyze(interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(result.Excluded).To(BeEmpty(), "Single 404 should not exclude any interactions")
}

func TestMonitorDeletion_TwoGETsThenConfirmation_NothingIsRemoved(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Two successful GETs followed by 404
	get1 := fake.NewInteraction(baseURL, "GET", 200)
	get2 := fake.NewInteraction(baseURL, "GET", 200)
	get404 := fake.NewInteraction(baseURL, "GET", 404)

	result1, err := monitor.Analyze(get1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Finished).To(BeFalse())

	result2, err := monitor.Analyze(get2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Finished).To(BeFalse())

	result3, err := monitor.Analyze(get404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result3.Finished).To(BeTrue())

	// With only 2 accumulated interactions, none should be excluded
	g.Expect(result3.Excluded).To(BeEmpty())
}

func TestMonitorDeletion_ThreeGETsThenConfirmation_MiddleIsRemoved(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Three successful GETs followed by 404
	get1 := fake.NewInteraction(baseURL, "GET", 200)
	get2 := fake.NewInteraction(baseURL, "GET", 200)
	get3 := fake.NewInteraction(baseURL, "GET", 200)
	get404 := fake.NewInteraction(baseURL, "GET", 404)

	result1, err := monitor.Analyze(get1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Finished).To(BeFalse())

	result2, err := monitor.Analyze(get2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Finished).To(BeFalse())

	result3, err := monitor.Analyze(get3)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result3.Finished).To(BeFalse())

	result4, err := monitor.Analyze(get404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result4.Finished).To(BeTrue())

	// First and last accumulated should remain, middle should be excluded
	g.Expect(result4.Excluded).To(HaveLen(1))
	g.Expect(result4.Excluded).To(ContainElement(get2))
}

func TestMonitorDeletion_MultipleMiddleGETs_AllMiddleAreRemoved(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Many successful GETs followed by 404
	interactions := make([]*fake.Interaction, 0, 10)
	for i := 0; i < 9; i++ {
		get := fake.NewInteraction(baseURL, "GET", 200)
		interactions = append(interactions, get)

		result, err := monitor.Analyze(get)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse())
	}

	get404 := fake.NewInteraction(baseURL, "GET", 404)
	interactions = append(interactions, get404)

	result, err := monitor.Analyze(get404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())

	// Verify exclusion pattern: all middle interactions should be excluded
	g.Expect(result.Excluded).To(HaveLen(7))
	for i := 1; i < 8; i++ {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]), "Middle GET %d should be excluded", i)
	}
}

func TestMonitorDeletion_DifferentURL_Ignored(t *testing.T) {
	g := NewWithT(t)
	monitoredURL := mustParseURL("https://api.example.com/resource/123")
	differentURL := mustParseURL("https://api.example.com/resource/456")
	monitor := generic.NewMonitorDeletion(monitoredURL)

	interaction := fake.NewInteraction(differentURL, "GET", 200)

	result, err := monitor.Analyze(interaction)

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
		"POST":   {method: "POST", statusCode: 201},
		"PUT":    {method: "PUT", statusCode: 200},
		"GET500": {method: "GET", statusCode: 500},
		"GET301": {method: "GET", statusCode: 301},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {

			g := NewWithT(t)
			baseURL := mustParseURL("https://api.example.com/resource/123")
			monitor := generic.NewMonitorDeletion(baseURL)

			// Start with some successful GETs
			get1 := fake.NewInteraction(baseURL, "GET", 200)
			result1, err := monitor.Analyze(get1)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result1.Finished).To(BeFalse())

			// Then a request that should abandon monitoring
			abandoningRequest := fake.NewInteraction(baseURL, c.method, c.statusCode)
			result2, err := monitor.Analyze(abandoningRequest)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result2.Finished).To(BeTrue())
			g.Expect(result2.Excluded).To(BeEmpty())
		})
	}
}

func TestMonitorDeletion_Various2xxStatusCodes_Accumulated(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Test various 2xx status codes
	statusCodes := []int{200, 201, 202, 204, 206}
	interactions := make([]*fake.Interaction, 0, len(statusCodes)+1)

	for _, code := range statusCodes {
		get := fake.NewInteraction(baseURL, "GET", code)
		interactions = append(interactions, get)

		result, err := monitor.Analyze(get)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse())
	}

	get404 := fake.NewInteraction(baseURL, "GET", 404)
	interactions = append(interactions, get404)

	result, err := monitor.Analyze(get404)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())

	// First and last accumulated should remain, middle should be excluded
	g.Expect(result.Excluded).To(HaveLen(3))
	for i := 1; i < len(statusCodes)-1; i++ {
		g.Expect(result.Excluded).To(ContainElement(interactions[i]))
	}
}

func TestMonitorDeletion_URLWithQueryParameters_MonitorsBaseURL(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	urlWithParams := mustParseURL("https://api.example.com/resource/123?param=value")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Interaction with query parameters should match base URL
	interaction := fake.NewInteraction(urlWithParams, "GET", 200)
	result, err := monitor.Analyze(interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
}

func TestMonitorDeletion_EmptyResult_WhenIgnoringInteraction(t *testing.T) {
	g := NewWithT(t)
	monitoredURL := mustParseURL("https://api.example.com/resource/123")
	differentURL := mustParseURL("https://api.example.com/other")
	monitor := generic.NewMonitorDeletion(monitoredURL)

	interaction := fake.NewInteraction(differentURL, "GET", 200)
	result, err := monitor.Analyze(interaction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}

// Helper functions

func mustParseURL(rawURL string) url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return *parsed
}
