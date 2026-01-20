package urltool

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/must"
)

func TestBaseURL_WithQueryAndFragment_StripsQueryAndFragment(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := must.ParseURL(t, "https://example.com/foo/bar?x=1#section")

	result := BaseURL(input)

	g.Expect(result.RawQuery).To(BeEmpty())
	g.Expect(result.Fragment).To(BeEmpty())
	g.Expect(result.String()).To(Equal("https://example.com/foo/bar"))
}

func TestBaseURL_WithDotSegments_CanonicalizesPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := must.ParseURL(t, "https://example.com/foo/../bar/./baz?skip=1#frag")

	result := BaseURL(input)

	g.Expect(result.String()).To(Equal("https://example.com/bar/baz"))
}

func TestSameBaseURL_DifferingOnlyOnQuery_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/resource?first=1")
	right := must.ParseURL(t, "https://example.com/resource?second=2")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_DifferingOnlyOnFragment_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/resource#part1")
	right := must.ParseURL(t, "https://example.com/resource#part2")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_DifferentPaths_DoesNotMatchOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/alpha")
	right := must.ParseURL(t, "https://example.com/beta")

	g.Expect(SameBaseURL(left, right)).To(BeFalse())
}

func TestSameBaseURL_SameCanonicalPaths_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/foo/../bar")
	right := must.ParseURL(t, "https://example.com/bar")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameBaseURL_IdenticalPaths_MatchesOnBase(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/foo/bar")
	right := must.ParseURL(t, "https://example.com/foo/bar")

	g.Expect(SameBaseURL(left, right)).To(BeTrue())
}

func TestSameURL_IdenticalURLs_Matches(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/foo/bar?x=1#section")
	right := must.ParseURL(t, "https://example.com/foo/bar?x=1#section")

	g.Expect(SameURL(left, right)).To(BeTrue())
}

func TestSameURL_DifferentQueryParameters_DoesNotMatch(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/resource?first=1")
	right := must.ParseURL(t, "https://example.com/resource?second=2")

	g.Expect(SameURL(left, right)).To(BeFalse())
}

func TestSameURL_DifferentFragments_DoesNotMatch(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/resource#part1")
	right := must.ParseURL(t, "https://example.com/resource#part2")

	g.Expect(SameURL(left, right)).To(BeFalse())
}

func TestSameURL_DifferentPaths_DoesNotMatch(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/alpha")
	right := must.ParseURL(t, "https://example.com/beta")

	g.Expect(SameURL(left, right)).To(BeFalse())
}

func TestSameURL_DifferentCanonicalPaths_DoesNotMatch(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/foo/../bar")
	right := must.ParseURL(t, "https://example.com/bar")

	g.Expect(SameURL(left, right)).To(BeFalse())
}

func TestSameURL_IdenticalPathsWithoutQueryOrFragment_Matches(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	left := must.ParseURL(t, "https://example.com/foo/bar")
	right := must.ParseURL(t, "https://example.com/foo/bar")

	g.Expect(SameURL(left, right)).To(BeTrue())
}
