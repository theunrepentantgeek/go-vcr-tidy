package generic

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
)

func TestDetectDeletion_SuccessfulDELETE_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	detector := NewDetectDeletion()

	// Successful DELETE should spawn a MonitorDeletion analyzer
	deleteInteraction := fake.NewInteraction(baseURL, "DELETE", 200)
	result, err := detector.Analyze(deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorDeletion{}))
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestDetectDeletion_Various2xxDELETEStatusCodes_SpawnsMonitor(t *testing.T) {
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
			baseURL := mustParseURL("https://api.example.com/resource/123")
			detector := NewDetectDeletion()

			deleteInteraction := fake.NewInteraction(baseURL, "DELETE", c.statusCode)
			result, err := detector.Analyze(deleteInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result.Spawn).To(HaveLen(1))
			g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorDeletion{}))
		})
	}
}

func TestDetectDeletion_FailedDELETE_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		"400 Bad Request":    {statusCode: 400},
		"401 Unauthorized":   {statusCode: 401},
		"403 Forbidden":      {statusCode: 403},
		"404 Not Found":      {statusCode: 404},
		"500 Server Error":   {statusCode: 500},
		"301 Redirect":       {statusCode: 301},
		"100 Informational":  {statusCode: 100},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := mustParseURL("https://api.example.com/resource/123")
			detector := NewDetectDeletion()

			deleteInteraction := fake.NewInteraction(baseURL, "DELETE", c.statusCode)
			result, err := detector.Analyze(deleteInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectDeletion_NonDELETEMethods_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method     string
		statusCode int
	}{
		"GET":     {method: "GET", statusCode: 200},
		"POST":    {method: "POST", statusCode: 201},
		"PUT":     {method: "PUT", statusCode: 200},
		"PATCH":   {method: "PATCH", statusCode: 200},
		"HEAD":    {method: "HEAD", statusCode: 200},
		"OPTIONS": {method: "OPTIONS", statusCode: 200},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := mustParseURL("https://api.example.com/resource/123")
			detector := NewDetectDeletion()

			interaction := fake.NewInteraction(baseURL, c.method, c.statusCode)
			result, err := detector.Analyze(interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectDeletion_MultipleDELETEs_SpawnsMultipleMonitors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	detector := NewDetectDeletion()

	// Different URLs being deleted
	url1 := mustParseURL("https://api.example.com/resource/123")
	url2 := mustParseURL("https://api.example.com/resource/456")
	url3 := mustParseURL("https://api.example.com/other/789")

	delete1 := fake.NewInteraction(url1, "DELETE", 200)
	delete2 := fake.NewInteraction(url2, "DELETE", 204)
	delete3 := fake.NewInteraction(url3, "DELETE", 202)

	result1, err := detector.Analyze(delete1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Spawn).To(HaveLen(1))

	result2, err := detector.Analyze(delete2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Spawn).To(HaveLen(1))

	result3, err := detector.Analyze(delete3)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result3.Spawn).To(HaveLen(1))

	// Each spawned monitor should be independent
	g.Expect(result1.Spawn[0]).ToNot(Equal(result2.Spawn[0]))
	g.Expect(result1.Spawn[0]).ToNot(Equal(result3.Spawn[0]))
	g.Expect(result2.Spawn[0]).ToNot(Equal(result3.Spawn[0]))
}

func TestDetectDeletion_NeverFinishes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	detector := NewDetectDeletion()

	// Process various interactions
	interactions := []*fake.Interaction{
		fake.NewInteraction(baseURL, "GET", 200),
		fake.NewInteraction(baseURL, "POST", 201),
		fake.NewInteraction(baseURL, "DELETE", 200),
		fake.NewInteraction(baseURL, "PUT", 200),
		fake.NewInteraction(baseURL, "DELETE", 404),
		fake.NewInteraction(baseURL, "GET", 404),
	}

	for _, inter := range interactions {
		result, err := detector.Analyze(inter)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse(), "DetectDeletion should never finish")
	}
}

func TestDetectDeletion_SpawnedMonitorHasCorrectURL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	detector := NewDetectDeletion()

	deleteInteraction := fake.NewInteraction(baseURL, "DELETE", 200)
	result, err := detector.Analyze(deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))

	// Verify the spawned monitor is configured for the correct URL
	monitor := result.Spawn[0].(*MonitorDeletion)
	g.Expect(monitor.baseURL).To(Equal(baseURL))
}

func TestDetectDeletion_EmptyResult_WhenNoAction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := mustParseURL("https://api.example.com/resource/123")
	detector := NewDetectDeletion()

	// Non-DELETE interaction
	getInteraction := fake.NewInteraction(baseURL, "GET", 200)
	result, err := detector.Analyze(getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
