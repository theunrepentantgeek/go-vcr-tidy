package interaction_test

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

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

// HasAnyMethod Tests

func TestHasAnyMethod(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		method          string
		statusCode      int
		methodsToCheck  []string
		expectedMatches bool
	}{
		"WithSingleMatchingMethod_ReturnsTrue": {
			method:          http.MethodGet,
			statusCode:      200,
			methodsToCheck:  []string{http.MethodGet},
			expectedMatches: true,
		},
		"WithMultipleMethodsFirstMatches_ReturnsTrue": {
			method:          http.MethodPost,
			statusCode:      201,
			methodsToCheck:  []string{http.MethodPost, http.MethodPut, http.MethodPatch},
			expectedMatches: true,
		},
		"WithMultipleMethodsMiddleMatches_ReturnsTrue": {
			method:          http.MethodPut,
			statusCode:      200,
			methodsToCheck:  []string{http.MethodPost, http.MethodPut, http.MethodPatch},
			expectedMatches: true,
		},
		"WithMultipleMethodsLastMatches_ReturnsTrue": {
			method:          http.MethodPatch,
			statusCode:      200,
			methodsToCheck:  []string{http.MethodPost, http.MethodPut, http.MethodPatch},
			expectedMatches: true,
		},
		"WithNoMatch_ReturnsFalse": {
			method:          http.MethodGet,
			statusCode:      200,
			methodsToCheck:  []string{http.MethodPost, http.MethodPut, http.MethodPatch},
			expectedMatches: false,
		},
		"WithEmptyMethodList_ReturnsFalse": {
			method:          http.MethodGet,
			statusCode:      200,
			methodsToCheck:  []string{},
			expectedMatches: false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://example.com/resource")
			i := fake.Interaction(baseURL, c.method, c.statusCode)

			result := interaction.HasAnyMethod(i, c.methodsToCheck...)

			g.Expect(result).To(Equal(c.expectedMatches))
		})
	}
}

// WasSuccessful Tests

func TestWasSuccessful(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		statusCode int
		expected   bool
	}{
		"With200_ReturnsTrue": {
			statusCode: 200,
			expected:   true,
		},
		"With201_ReturnsTrue": {
			statusCode: 201,
			expected:   true,
		},
		"With204_ReturnsTrue": {
			statusCode: 204,
			expected:   true,
		},
		"With299_ReturnsTrue": {
			statusCode: 299,
			expected:   true,
		},
		"With199_ReturnsFalse": {
			statusCode: 199,
			expected:   false,
		},
		"With300_ReturnsFalse": {
			statusCode: 300,
			expected:   false,
		},
		"With404_ReturnsFalse": {
			statusCode: 404,
			expected:   false,
		},
		"With500_ReturnsFalse": {
			statusCode: 500,
			expected:   false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			baseURL := mustParseURL("https://example.com/resource")
			i := fake.Interaction(baseURL, http.MethodGet, c.statusCode)

			result := interaction.WasSuccessful(i)

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
