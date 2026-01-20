package azure

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestMonitorLongRunningOperation_Rewire_AddsLocationHeader_WhenURLsDiffer(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	// Create two interactions with different URLs
	priorURL := must.ParseURL(t, "https://management.azure.com/operations/1")
	nextURL := must.ParseURL(t, "https://management.azure.com/operations/2")

	prior := fake.Interaction(priorURL, http.MethodGet, 200)
	next := fake.Interaction(nextURL, http.MethodGet, 200)

	// Verify no Location header initially
	_, ok := prior.Response().Header("Location")
	g.Expect(ok).To(BeFalse())

	// Rewire should add Location header
	monitor.rewire(prior, next)

	location, ok := prior.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location).To(Equal(nextURL.String()))
}

func TestMonitorLongRunningOperation_Rewire_RemovesLocationHeader_WhenURLsSame(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	// Create two interactions with the same URL
	sameURL := must.ParseURL(t, "https://management.azure.com/operations/1")

	prior := fake.Interaction(sameURL, http.MethodGet, 200)
	next := fake.Interaction(sameURL, http.MethodGet, 200)

	// Add a Location header to the prior interaction
	prior.SetResponseHeader("Location", "https://somewhere.else.com")

	// Verify Location header exists
	_, ok := prior.Response().Header("Location")
	g.Expect(ok).To(BeTrue())

	// Rewire should remove Location header
	monitor.rewire(prior, next)

	_, ok = prior.Response().Header("Location")
	g.Expect(ok).To(BeFalse())
}

func TestMonitorLongRunningOperation_Rewire_KeepsExistingLocationHeader_WhenURLsDiffer(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	priorURL := must.ParseURL(t, "https://management.azure.com/operations/1")
	nextURL := must.ParseURL(t, "https://management.azure.com/operations/2")

	prior := fake.Interaction(priorURL, http.MethodGet, 200)
	next := fake.Interaction(nextURL, http.MethodGet, 200)

	// Add a Location header to the prior interaction
	existingLocation := "https://existing.location.com"
	prior.SetResponseHeader("Location", existingLocation)

	// Rewire should keep the existing Location header
	monitor.rewire(prior, next)

	location, ok := prior.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location).To(Equal(existingLocation))
}

func TestMonitorLongRunningOperation_Rewire_HandlesQueryParameters(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	// URLs with different query parameters are considered different
	priorURL := must.ParseURL(t, "https://management.azure.com/operations/1?t=123")
	nextURL := must.ParseURL(t, "https://management.azure.com/operations/1?t=456")

	prior := fake.Interaction(priorURL, http.MethodGet, 200)
	next := fake.Interaction(nextURL, http.MethodGet, 200)

	// Rewire should add Location header because URLs differ
	monitor.rewire(prior, next)

	location, ok := prior.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location).To(Equal(nextURL.String()))
}

func TestMonitorLongRunningOperation_RewireSequence_LinksAllInteractions(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	// Create a sequence of interactions with different URLs
	url1 := must.ParseURL(t, "https://management.azure.com/operations/1")
	url2 := must.ParseURL(t, "https://management.azure.com/operations/2")
	url3 := must.ParseURL(t, "https://management.azure.com/operations/3")

	i1 := fake.Interaction(url1, http.MethodGet, 200)
	i2 := fake.Interaction(url2, http.MethodGet, 200)
	i3 := fake.Interaction(url3, http.MethodGet, 200)

	interactions := []interaction.Interface{i1, i2, i3}

	// Rewire the sequence
	monitor.rewireSequence(interactions)

	// Verify i1 has Location header pointing to i2
	location1, ok := i1.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location1).To(Equal(url2.String()))

	// Verify i2 has Location header pointing to i3
	location2, ok := i2.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location2).To(Equal(url3.String()))

	// Verify i3 has no Location header (it's the last one)
	_, ok = i3.Response().Header("Location")
	g.Expect(ok).To(BeFalse())
}

func TestMonitorLongRunningOperation_RewireSequence_HandlesSingleInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	url1 := must.ParseURL(t, "https://management.azure.com/operations/1")
	i1 := fake.Interaction(url1, http.MethodGet, 200)

	interactions := []interaction.Interface{i1}

	// Rewire should handle single interaction gracefully
	monitor.rewireSequence(interactions)

	// No Location header should be added
	_, ok := i1.Response().Header("Location")
	g.Expect(ok).To(BeFalse())
}

func TestMonitorLongRunningOperation_RewireSequence_HandlesMixedURLs(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	monitor := &MonitorAzureLongRunningOperation{}

	// Create sequence where some URLs are same, some different
	url1 := must.ParseURL(t, "https://management.azure.com/operations/1")
	url2 := must.ParseURL(t, "https://management.azure.com/operations/1") // Same as url1
	url3 := must.ParseURL(t, "https://management.azure.com/operations/2") // Different

	i1 := fake.Interaction(url1, http.MethodGet, 200)
	i2 := fake.Interaction(url2, http.MethodGet, 200)
	i3 := fake.Interaction(url3, http.MethodGet, 200)

	// Add a Location header to i1 that should be removed (since i1 and i2 have same URL)
	i1.SetResponseHeader("Location", "https://old.location.com/1")

	interactions := []interaction.Interface{i1, i2, i3}

	// Rewire the sequence
	monitor.rewireSequence(interactions)

	// i1 and i2 have same URL, so i1 should have no Location header
	_, ok := i1.Response().Header("Location")
	g.Expect(ok).To(BeFalse())

	// i2 and i3 have different URLs, so i2 should have Location header pointing to i3
	location2, ok := i2.Response().Header("Location")
	g.Expect(ok).To(BeTrue())
	g.Expect(location2).To(Equal(url3.String()))

	// i3 is the last one, so no Location header
	_, ok = i3.Response().Header("Location")
	g.Expect(ok).To(BeFalse())
}

func TestMonitorLongRunningOperation_RewireSequence_EmptySlice(t *testing.T) {
	t.Parallel()

	monitor := &MonitorAzureLongRunningOperation{}

	// Rewire should handle empty slice gracefully
	monitor.rewireSequence([]interaction.Interface{})

	// No panic should occur - test passes if it completes
}
