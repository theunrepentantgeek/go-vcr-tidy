package generic

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/neilotoole/slogt"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestDetectDeferredCreation_GET404_SpawnsMonitor(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeferredCreation()
	log := slogt.New(t)

	getInteraction := fake.Interaction(baseURL, http.MethodGet, 404)
	result, err := detector.Analyze(log, getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Finished).To(BeFalse())
	g.Expect(result.Spawn).To(HaveLen(1))
	g.Expect(result.Spawn[0]).To(BeAssignableToTypeOf(&MonitorDeferredCreation{}))
	g.Expect(result.Excluded).To(BeEmpty())
}

func TestDetectDeferredCreation_Non404StatusCodes_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
	}{
		"200 OK":              {statusCode: 200},
		"201 Created":         {statusCode: 201},
		"400 Bad Request":     {statusCode: 400},
		"401 Unauthorized":    {statusCode: 401},
		"403 Forbidden":       {statusCode: 403},
		"500 Internal Server": {statusCode: 500},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			detector := NewDetectDeferredCreation()
			log := slogt.New(t)

			getInteraction := fake.Interaction(baseURL, http.MethodGet, c.statusCode)
			result, err := detector.Analyze(log, getInteraction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

func TestDetectDeferredCreation_NonGETMethodsWith404_DoesNotSpawn(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method string
	}{
		"POST":    {method: http.MethodPost},
		"PUT":     {method: http.MethodPut},
		"DELETE":  {method: http.MethodDelete},
		"PATCH":   {method: http.MethodPatch},
		"HEAD":    {method: http.MethodHead},
		"OPTIONS": {method: http.MethodOptions},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
			detector := NewDetectDeferredCreation()
			log := slogt.New(t)

			interaction := fake.Interaction(baseURL, c.method, 404)
			result, err := detector.Analyze(log, interaction)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(analyzer.Result{}))
		})
	}
}

//nolint:dupl // Similar structure to TestDetectDeletion_MultipleDELETEs_SpawnsMultipleMonitors
func TestDetectDeferredCreation_MultipleGET404s_SpawnsMultipleMonitors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	detector := NewDetectDeferredCreation()
	log := slogt.New(t)

	url1 := must.ParseURL(t, "https://api.example.com/resource/123")
	url2 := must.ParseURL(t, "https://api.example.com/resource/456")
	url3 := must.ParseURL(t, "https://api.example.com/other/789")

	get1 := fake.Interaction(url1, http.MethodGet, 404)
	get2 := fake.Interaction(url2, http.MethodGet, 404)
	get3 := fake.Interaction(url3, http.MethodGet, 404)

	result1, err := detector.Analyze(log, get1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result1.Spawn).To(HaveLen(1))

	result2, err := detector.Analyze(log, get2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result2.Spawn).To(HaveLen(1))

	result3, err := detector.Analyze(log, get3)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result3.Spawn).To(HaveLen(1))

	g.Expect(result1.Spawn[0]).ToNot(Equal(result2.Spawn[0]))
	g.Expect(result1.Spawn[0]).ToNot(Equal(result3.Spawn[0]))
	g.Expect(result2.Spawn[0]).ToNot(Equal(result3.Spawn[0]))
}

func TestDetectDeferredCreation_NeverFinishes(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeferredCreation()
	log := slogt.New(t)

	interactions := []*fake.TestInteraction{
		fake.Interaction(baseURL, http.MethodGet, 200),
		fake.Interaction(baseURL, http.MethodPost, 201),
		fake.Interaction(baseURL, http.MethodGet, 404),
		fake.Interaction(baseURL, http.MethodPut, 200),
		fake.Interaction(baseURL, http.MethodDelete, 204),
		fake.Interaction(baseURL, http.MethodGet, 500),
	}

	for _, inter := range interactions {
		result, err := detector.Analyze(log, inter)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result.Finished).To(BeFalse(), "DetectDeferredCreation should never finish")
	}
}

func TestDetectDeferredCreation_EmptyResult_WhenNoAction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	baseURL := must.ParseURL(t, "https://api.example.com/resource/123")
	detector := NewDetectDeferredCreation()
	log := slogt.New(t)

	getInteraction := fake.Interaction(baseURL, http.MethodGet, 200)
	result, err := detector.Analyze(log, getInteraction)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(analyzer.Result{}))
}
