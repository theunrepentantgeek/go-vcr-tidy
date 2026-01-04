package generic

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestDetectDeletion_SuccessfulDELETE_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeletion()
	log := newTestLogger(t)

	// Successful DELETE should spawn a MonitorDeletion analyzer
	deleteInteraction := fake.Interaction(baseURL, http.MethodDelete, 200)
	result, err := detector.Analyze(log, deleteInteraction)

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
			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			detector := NewDetectDeletion()
			log := newTestLogger(t)

			deleteInteraction := fake.Interaction(baseURL, http.MethodDelete, c.statusCode)
			result, err := detector.Analyze(log, deleteInteraction)

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
		"400 Bad Request":   {statusCode: 400},
		"401 Unauthorized":  {statusCode: 401},
		"403 Forbidden":     {statusCode: 403},
		"404 Not Found":     {statusCode: 404},
		"500 Server Error":  {statusCode: 500},
		"301 Redirect":      {statusCode: 301},
		"100 Informational": {statusCode: 100},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			detector := NewDetectDeletion()
			log := newTestLogger(t)

			deleteInteraction := fake.Interaction(baseURL, http.MethodDelete, c.statusCode)
			result, err := detector.Analyze(log, deleteInteraction)

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
		"GET":     {method: http.MethodGet, statusCode: 200},
		"POST":    {method: http.MethodPost, statusCode: 201},
		"PUT":     {method: http.MethodPut, statusCode: 200},
		"PATCH":   {method: http.MethodPatch, statusCode: 200},
		"HEAD":    {method: http.MethodHead, statusCode: 200},
		"OPTIONS": {method: http.MethodOptions, statusCode: 200},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			detector := NewDetectDeletion()
			log := newTestLogger(t)

			interaction := fake.Interaction(baseURL, c.method, c.statusCode)
			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectDeletion_MultipleDELETEs_SpawnsMultipleMonitors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	detector := NewDetectDeletion()
	log := newTestLogger(t)

	// Different URLs being deleted
	url1 := must.ParseURL(t, "https://api.example.com/resource/123")
	url2 := must.ParseURL(t, "https://api.example.com/resource/456")
	url3 := must.ParseURL(t, "https://api.example.com/other/789")

	delete1 := fake.Interaction(url1, http.MethodDelete, 200)
	delete2 := fake.Interaction(url2, http.MethodDelete, 204)
	delete3 := fake.Interaction(url3, http.MethodDelete, 202)

	result1, err := detector.Analyze(log, delete1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Spawn).To(HaveLen(1))

	result2, err := detector.Analyze(log, delete2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Spawn).To(HaveLen(1))

	result3, err := detector.Analyze(log, delete3)
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
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeletion()
	log := newTestLogger(t)

	// Process various interactions
	interactions := []*fake.TestInteraction{
		fake.Interaction(baseURL, http.MethodGet, 200),
		fake.Interaction(baseURL, http.MethodPost, 201),
		fake.Interaction(baseURL, http.MethodDelete, 200),
		fake.Interaction(baseURL, http.MethodPut, 200),
		fake.Interaction(baseURL, http.MethodDelete, 404),
		fake.Interaction(baseURL, http.MethodGet, 404),
	}

	for _, inter := range interactions {
		result, err := detector.Analyze(log, inter)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse(), "DetectDeletion should never finish")
	}
}

func TestDetectDeletion_SpawnedMonitorHasCorrectURL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeletion()
	log := newTestLogger(t)

	deleteInteraction := fake.Interaction(baseURL, http.MethodDelete, 200)
	result, err := detector.Analyze(log, deleteInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorDeletion{}))
	// Verify the spawned monitor is configured for the correct URL
	if m, ok := result.Spawn[0].(*MonitorDeletion); ok {
		g.Expect(m.baseURL).To(Equal(baseURL))
	}
}

func TestDetectDeletion_EmptyResult_WhenNoAction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeletion()
	log := newTestLogger(t)

	// Non-DELETE interaction
	getInteraction := fake.Interaction(baseURL, http.MethodGet, 200)
	result, err := detector.Analyze(log, getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
