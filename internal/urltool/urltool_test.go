package urltool

import (
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
)

func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	return *parsed
}

func TestBaseURL_WithQueryAndFragment_StripsQueryAndFragment(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := mustParseURL(t, "https://example.com/foo/bar?x=1#section")

	result := BaseURL(input)

	g.Expect(result.RawQuery).To(BeEmpty())
	g.Expect(result.Fragment).To(BeEmpty())
	g.Expect(result.String()).To(Equal("https://example.com/foo/bar"))
}

func TestBaseURL_WithDotSegments_CanonicalizesPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := mustParseURL(t, "https://example.com/foo/../bar/./baz?skip=1#frag")

	result := BaseURL(input)

	g.Expect(result.String()).To(Equal("https://example.com/bar/baz"))
}

func TestSameBaseURL_DifferingOnlyOnQuery_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := mustParseURL(t, "https://example.com/resource?first=1")
	right := mustParseURL(t, "https://example.com/resource?second=2")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_DifferingOnlyOnFragment_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := mustParseURL(t, "https://example.com/resource#part1")
	right := mustParseURL(t, "https://example.com/resource#part2")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_DifferentPaths_DoesNotMatchOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := mustParseURL(t, "https://example.com/alpha")
	right := mustParseURL(t, "https://example.com/beta")

	g.Expect(SameBaseURL(left, right)).To(BeFalse())
}

func TestSameBaseURL_SameCanonicalPaths_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := mustParseURL(t, "https://example.com/foo/../bar")
	right := mustParseURL(t, "https://example.com/bar")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_IdenticalPaths_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := mustParseURL(t, "https://example.com/foo/bar")
	right := mustParseURL(t, "https://example.com/foo/bar")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}
