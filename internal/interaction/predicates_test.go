package interaction

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/google/uuid"
)

// Test mock implementations

type mockInteraction struct {
	request  mockRequest
	response mockResponse
}

func (m *mockInteraction) ID() uuid.UUID {
	return uuid.New()
}

func (m *mockInteraction) Request() Request {
	return &m.request
}

func (m *mockInteraction) Response() Response {
	return &m.response
}

type mockRequest struct {
	method string
}

func (m *mockRequest) FullURL() url.URL {
	return url.URL{}
}

func (m *mockRequest) BaseURL() url.URL {
	return url.URL{}
}

func (m *mockRequest) Method() string {
	return m.method
}

type mockResponse struct {
	statusCode int
}

func (m *mockResponse) StatusCode() int {
	return m.statusCode
}

func (m *mockResponse) Header(name string) (string, bool) {
	return "", false
}

func (m *mockResponse) Body() []byte {
	return nil
}

func newMockInteraction(method string, statusCode int) *mockInteraction {
	return &mockInteraction{
		request: mockRequest{
			method: method,
		},
		response: mockResponse{
			statusCode: statusCode,
		},
	}
}

// HasMethod Tests

func TestHasMethod_WithMatchingMethod_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasMethod(i, http.MethodGet)

	g.Expect(result).To(BeTrue())
}

func TestHasMethod_WithNonMatchingMethod_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasMethod(i, http.MethodPost)

	g.Expect(result).To(BeFalse())
}

func TestHasMethod_WithDifferentCase_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasMethod(i, "get")

	g.Expect(result).To(BeFalse())
}

// HasAnyMethod Tests

func TestHasAnyMethod_WithSingleMatchingMethod_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasAnyMethod(i, http.MethodGet)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsFirstMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodPost, 201)

	result := HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsMiddleMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodPut, 200)

	result := HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithMultipleMethodsLastMatches_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodPatch, 200)

	result := HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeTrue())
}

func TestHasAnyMethod_WithNoMatch_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasAnyMethod(i, http.MethodPost, http.MethodPut, http.MethodPatch)

	g.Expect(result).To(BeFalse())
}

func TestHasAnyMethod_WithEmptyMethodList_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := HasAnyMethod(i)

	g.Expect(result).To(BeFalse())
}

// WasSuccessful Tests

func TestWasSuccessful_With200_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 200)

	result := WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With201_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodPost, 201)

	result := WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With204_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodDelete, 204)

	result := WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With299_ReturnsTrue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 299)

	result := WasSuccessful(i)

	g.Expect(result).To(BeTrue())
}

func TestWasSuccessful_With199_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 199)

	result := WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With300_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 300)

	result := WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With404_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 404)

	result := WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}

func TestWasSuccessful_With500_ReturnsFalse(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	i := newMockInteraction(http.MethodGet, 500)

	result := WasSuccessful(i)

	g.Expect(result).To(BeFalse())
}
