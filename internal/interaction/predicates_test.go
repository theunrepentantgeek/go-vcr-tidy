package interaction_test

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

func mustParseURL(raw string) url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return *parsed
}

// HasMethod Tests

func TestHasMethod_WithMatchingMethod_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasMethod(i, http.MethodGet)

	g.Expect(result).To(BeTrue())
}

func TestHasMethod_WithNonMatchingMethod_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasMethod(i, http.MethodPost)

	g.Expect(result).To(BeFalse())
}

func TestHasMethod_WithDifferentCase_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasMethod(i, "get")

	g.Expect(result).To(BeFalse())
}

// HasAnyMethod Tests

func TestHasAnyMethod_WithSingleMatchingMethod_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasAnyMethod(i, http.MethodGet)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsFirstMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodPost, 201)

	result := interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsMiddleMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodPut, 200)

	result := interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsLastMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodPatch, 200)

	result := interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithNoMatch_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeFalse())
}

func TestHasAnyMethod_WithEmptyMethodList_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.HasAnyMethod(i)

	g.Expect(result).To(BeFalse())
}

// WasSuccessful Tests

func TestWasSuccessful_With200_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 200)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With201_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodPost, 201)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With204_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodDelete, 204)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With299_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 299)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With199_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 199)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With300_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 300)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With404_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 404)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With500_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://example.com/resource")
	i := fake.Interaction(baseURL, http.MethodGet, 500)

	result := interaction.WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}
