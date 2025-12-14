package generic_test

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"
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
	g.Expect(interaction.IsMarkedForRemoval()).To(BeFalse(), "Single 404 should not be marked for removal")
}

func TestMonitorDeletion_TwoGETsThenConfirmation_OnlyConfirmationRemains(t *testing.T) {
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

	// With only 2 accumulated interactions, none should be marked for removal
	g.Expect(get1.IsMarkedForRemoval()).To(BeFalse())
	g.Expect(get2.IsMarkedForRemoval()).To(BeFalse())
	g.Expect(get404.IsMarkedForRemoval()).To(BeFalse())
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

	// First and last accumulated should remain, middle should be removed
	g.Expect(get1.IsMarkedForRemoval()).To(BeFalse(), "First GET should not be removed")
	g.Expect(get2.IsMarkedForRemoval()).To(BeTrue(), "Middle GET should be removed")
	g.Expect(get3.IsMarkedForRemoval()).To(BeFalse(), "Last GET should not be removed")
	g.Expect(get404.IsMarkedForRemoval()).To(BeFalse(), "Confirmation 404 should not be removed")
}

func TestMonitorDeletion_MultipleMiddleGETs_AllMiddleAreRemoved(t *testing.T) {
	g := NewWithT(t)
	goldie := goldie.New(t)
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

	// Verify removal pattern
	g.Expect(interactions[0].IsMarkedForRemoval()).To(BeFalse(), "First GET should not be removed")
	for i := 1; i < 8; i++ {
		g.Expect(interactions[i].IsMarkedForRemoval()).To(BeTrue(), "Middle GET %d should be removed", i)
	}
	g.Expect(interactions[8].IsMarkedForRemoval()).To(BeFalse(), "Last GET should not be removed")
	g.Expect(interactions[9].IsMarkedForRemoval()).To(BeFalse(), "Confirmation 404 should not be removed")

	// Use goldie to snapshot the state
	output := formatInteractions(interactions)
	goldie.Assert(t, "multiple_middle_gets", []byte(output))
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
	g.Expect(interaction.IsMarkedForRemoval()).To(BeFalse())
}

func TestMonitorDeletion_NonGETMethod_AbandonMonitoring(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Start with some successful GETs
	get1 := fake.NewInteraction(baseURL, "GET", 200)
	result1, err := monitor.Analyze(get1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Finished).To(BeFalse())

	// Then a POST should abandon monitoring
	post := fake.NewInteraction(baseURL, "POST", 201)
	result2, err := monitor.Analyze(post)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Finished).To(BeTrue())
	g.Expect(get1.IsMarkedForRemoval()).To(BeFalse())
	g.Expect(post.IsMarkedForRemoval()).To(BeFalse())
}

func TestMonitorDeletion_PUTMethod_AbandonMonitoring(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	put := fake.NewInteraction(baseURL, "PUT", 200)
	result, err := monitor.Analyze(put)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
}

func TestMonitorDeletion_GETWith500Status_AbandonMonitoring(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	get500 := fake.NewInteraction(baseURL, "GET", 500)
	result, err := monitor.Analyze(get500)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(get500.IsMarkedForRemoval()).To(BeFalse())
}

func TestMonitorDeletion_GETWith301Status_AbandonMonitoring(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	get301 := fake.NewInteraction(baseURL, "GET", 301)
	result, err := monitor.Analyze(get301)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(get301.IsMarkedForRemoval()).To(BeFalse())
}

func TestMonitorDeletion_Various2xxStatusCodes_Accumulated(t *testing.T) {
	g := NewWithT(t)
	goldie := goldie.New(t)
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

	// First and last accumulated should remain, middle should be removed
	g.Expect(interactions[0].IsMarkedForRemoval()).To(BeFalse())
	for i := 1; i < len(statusCodes)-1; i++ {
		g.Expect(interactions[i].IsMarkedForRemoval()).To(BeTrue())
	}
	g.Expect(interactions[len(statusCodes)-1].IsMarkedForRemoval()).To(BeFalse())
	g.Expect(get404.IsMarkedForRemoval()).To(BeFalse())

	output := formatInteractions(interactions)
	goldie.Assert(t, "various_2xx_codes", []byte(output))
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

func TestMonitorDeletion_CompleteSequence_GoldenFileSnapshot(t *testing.T) {
	g := NewWithT(t)
	goldie := goldie.New(t)
	baseURL := mustParseURL("https://api.example.com/resources/delete-me")
	monitor := generic.NewMonitorDeletion(baseURL)

	interactions := []*fake.Interaction{
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "GET", 404),
	}

	for i, interaction := range interactions {
		result, err := monitor.Analyze(interaction)
		g.Expect(err).ToNot(HaveOccurred())
		if i < len(interactions)-1 {
			g.Expect(result.Finished).To(BeFalse())
		} else {
			g.Expect(result.Finished).To(BeTrue())
		}
	}

	output := formatInteractions(interactions)
	goldie.Assert(t, "complete_sequence", []byte(output))
}

func TestMonitorDeletion_NoAccumulatedGETs_404StillFinishes(t *testing.T) {
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	monitor := generic.NewMonitorDeletion(baseURL)

	// Immediate 404 without any accumulated GETs
	get404 := fake.NewInteraction(baseURL, "GET", 404)
	result, err := monitor.Analyze(get404)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeTrue())
	g.Expect(get404.IsMarkedForRemoval()).To(BeFalse())
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

func formatInteractions(interactions []*fake.Interaction) string {
	result := ""
	for i, interaction := range interactions {
		if i > 0 {
			result += "\n"
		}
		result += interaction.String()
	}
	return result
}
