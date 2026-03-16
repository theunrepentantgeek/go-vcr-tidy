# Deferred Creation Polling Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a generic detector/monitor pair that removes intermediate 404 GET responses when a client polls waiting for a resource to be created.

**Architecture:** Two new types in `internal/generic/` mirroring the existing `DetectDeletion`/`MonitorDeletion` pair. `DetectDeferredCreation` watches for GET→404 and spawns `MonitorDeferredCreation`, which accumulates 404 GETs and excludes the middle ones when a GET→2xx confirms creation.

**Tech Stack:** Go, gomega (assertions), slogt (test logging), existing `analyzer`, `interaction`, `fake`, `must`, and `urltool` packages.

---

## Chunk 1: DetectDeferredCreation

### Task 1: Write DetectDeferredCreation detector tests

**Files:**
- Create: `internal/generic/detect_deferred_creation_test.go`

- [ ] **Step 1: Create test file with initial test — GET 404 spawns monitor**

```go
package generic

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/neilotoole/slogt"

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /workspaces/go-vcr-tidy && go test ./internal/generic/ -run TestDetectDeferredCreation_GET404_SpawnsMonitor -v`
Expected: FAIL — `NewDetectDeferredCreation` and `MonitorDeferredCreation` not defined.

- [ ] **Step 3: Add remaining detector tests to the same file**

Append the following tests after the first test:

```go
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
```

Note: `TestDetectDeferredCreation_Non404StatusCodes_DoesNotSpawn` needs an import for `analyzer` package. Add `"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"` to the imports block.

### Task 2: Implement DetectDeferredCreation

**Files:**
- Create: `internal/generic/detect_deferred_creation.go`

- [ ] **Step 1: Create the detector implementation**

```go
package generic

import (
	"log/slog"
	"net/http"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// DetectDeferredCreation is an analyzer for detecting polling for resource creation.
// It watches for GET requests that return 404 (Not Found) and spawns a MonitorDeferredCreation
// analyzer to track the subsequent GET requests until the resource is created.
type DetectDeferredCreation struct{}

var _ analyzer.Interface = &DetectDeferredCreation{}

// NewDetectDeferredCreation creates a new DetectDeferredCreation analyzer.
func NewDetectDeferredCreation() *DetectDeferredCreation {
	return &DetectDeferredCreation{}
}

// Analyze processes another interaction in the sequence.
func (*DetectDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()

	if interaction.HasMethod(i, http.MethodGet) && i.Response().StatusCode() == http.StatusNotFound {
		log.Debug(
			"Found GET 404 to monitor for deferred creation",
			"url", reqURL.String(),
		)

		monitor := NewMonitorDeferredCreation(i)

		return analyzer.Spawn(monitor), nil
	}

	return analyzer.Result{}, nil
}
```

- [ ] **Step 2: Create a stub MonitorDeferredCreation so the detector tests compile**

Create a minimal stub in `internal/generic/monitor_deferred_creation.go`:

```go
package generic

import (
	"log/slog"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// MonitorDeferredCreation is an analyzer for tracking polling for resource creation.
type MonitorDeferredCreation struct{}

// NewMonitorDeferredCreation creates a new MonitorDeferredCreation analyzer.
func NewMonitorDeferredCreation(
	firstInteraction interaction.Interface,
) *MonitorDeferredCreation {
	return &MonitorDeferredCreation{}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	return analyzer.Result{}, nil
}
```

- [ ] **Step 3: Run all detector tests to verify they pass**

Run: `cd /workspaces/go-vcr-tidy && go test ./internal/generic/ -run TestDetectDeferredCreation -v`
Expected: All PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/generic/detect_deferred_creation.go internal/generic/detect_deferred_creation_test.go internal/generic/monitor_deferred_creation.go
git commit -m "feat: add DetectDeferredCreation detector with tests"
```

## Chunk 2: MonitorDeferredCreation

### Task 3: Write MonitorDeferredCreation monitor tests

**Files:**
- Create: `internal/generic/monitor_deferred_creation_test.go`

- [ ] **Step 1: Create test file with all monitor tests**

```go
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

	for i := 0; i < 7; i++ {
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /workspaces/go-vcr-tidy && go test ./internal/generic/ -run TestMonitorDeferredCreation -v`
Expected: FAIL — `MonitorDeferredCreation` stub doesn't implement the full logic.

### Task 4: Implement MonitorDeferredCreation

**Files:**
- Modify: `internal/generic/monitor_deferred_creation.go` (replace stub)

- [ ] **Step 1: Replace the stub with the full implementation**

Replace the entire contents of `internal/generic/monitor_deferred_creation.go` with:

```go
package generic

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// MonitorDeferredCreation is an analyzer for tracking polling for resource creation.
// It watches for an uninterrupted sequence of GET requests returning 404 (Not Found) to a specific URL,
// followed by a GET that returns a 2xx status (indicating the resource has been created).
// Once creation is confirmed, the analyzer marks itself as Finished.
// All 404 GET requests are accumulated, and when the 2xx is seen, the analyzer indicates all but
// the first and last 404 are removable.
// If any other requests to that URL are seen (e.g. a POST or PUT), or if a GET returns a non-404
// and non-2xx status code, the analyzer abandons monitoring and marks itself as Finished.
type MonitorDeferredCreation struct {
	baseURL      *url.URL
	interactions []interaction.Interface
}

var _ analyzer.Interface = (*MonitorDeferredCreation)(nil)

// NewMonitorDeferredCreation creates a new MonitorDeferredCreation analyzer.
// firstInteraction is the initial GET→404 that triggered the detector.
func NewMonitorDeferredCreation(
	firstInteraction interaction.Interface,
) *MonitorDeferredCreation {
	return &MonitorDeferredCreation{
		baseURL:      firstInteraction.Request().BaseURL(),
		interactions: []interaction.Interface{firstInteraction},
	}
}

// Analyze processes another interaction in the sequence.
func (m *MonitorDeferredCreation) Analyze(
	log *slog.Logger,
	i interaction.Interface,
) (analyzer.Result, error) {
	reqURL := i.Request().BaseURL()
	statusCode := i.Response().StatusCode()

	switch {
	case !urltool.SameBaseURL(reqURL, m.baseURL):
		// Not the URL we're monitoring, ignore.
		return analyzer.Result{}, nil

	case interaction.HasMethod(i, http.MethodGet) && interaction.WasSuccessful(i):
		return m.creationConfirmed(log)

	case interaction.HasMethod(i, http.MethodGet) && statusCode == http.StatusNotFound:
		// Accumulate this 404 GET request.
		m.interactions = append(m.interactions, i)

		return analyzer.Result{}, nil

	case interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch):
		// Resource has changed, abandon monitoring.
		log.Debug(
			"Abandoning deferred creation monitor, resource changed",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
		)

		return analyzer.Finished(), nil

	default:
		// Unexpected method or status code, abandon monitoring.
		log.Debug(
			"Abandoning deferred creation monitor due to unexpected request",
			"url", m.baseURL.String(),
			"method", i.Request().Method(),
			"statusCode", statusCode,
		)

		return analyzer.Finished(), nil
	}
}

// creationConfirmed handles the confirmation of creation via a 2xx GET response.
func (m *MonitorDeferredCreation) creationConfirmed(
	log *slog.Logger,
) (analyzer.Result, error) {
	if len(m.interactions) < 3 {
		// Not enough intermediate interactions to exclude.
		log.Debug(
			"Short deferred creation monitor, nothing to exclude",
			"url", m.baseURL.String(),
		)

		return analyzer.Finished(), nil
	}

	log.Debug(
		"Long deferred creation found, excluding intermediate 404s",
		"url", m.baseURL.String(),
		"removed", len(m.interactions)-2,
	)

	excluded := m.interactions[1 : len(m.interactions)-1]

	return analyzer.FinishedWithExclusions(excluded...), nil
}
```

- [ ] **Step 2: Run all monitor tests to verify they pass**

Run: `cd /workspaces/go-vcr-tidy && go test ./internal/generic/ -run TestMonitorDeferredCreation -v`
Expected: All PASS.

- [ ] **Step 3: Run all generic package tests to verify nothing is broken**

Run: `cd /workspaces/go-vcr-tidy && go test ./internal/generic/ -v`
Expected: All PASS (both deletion and deferred creation tests).

- [ ] **Step 4: Run linter**

Run: `cd /workspaces/go-vcr-tidy && task lint`
Expected: No new lint errors.

- [ ] **Step 5: Commit**

```bash
git add internal/generic/monitor_deferred_creation.go internal/generic/monitor_deferred_creation_test.go
git commit -m "feat: add MonitorDeferredCreation monitor with tests"
```

- [ ] **Step 6: Run full test suite**

Run: `cd /workspaces/go-vcr-tidy && task test`
Expected: All tests pass.
