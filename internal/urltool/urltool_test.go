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

func TestSameURL(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		left     string
		right    string
		expected bool
	}{
		"IdenticalURLs_Matches": {
			left:     "https://example.com/foo/bar?x=1#section",
			right:    "https://example.com/foo/bar?x=1#section",
			expected: true,
		},
		"DifferentQueryParameters_DoesNotMatch": {
			left:     "https://example.com/resource?first=1",
			right:    "https://example.com/resource?second=2",
			expected: false,
		},
		"DifferentFragments_DoesNotMatch": {
			left:     "https://example.com/resource#part1",
			right:    "https://example.com/resource#part2",
			expected: false,
		},
		"DifferentPaths_DoesNotMatch": {
			left:     "https://example.com/alpha",
			right:    "https://example.com/beta",
			expected: false,
		},
		"DifferentCanonicalPaths_DoesNotMatch": {
			left:     "https://example.com/foo/../bar",
			right:    "https://example.com/bar",
			expected: false,
		},
		"IdenticalPathsWithoutQueryOrFragment_Matches": {
			left:     "https://example.com/foo/bar",
			right:    "https://example.com/foo/bar",
			expected: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			left := must.ParseURL(t, c.left)
			right := must.ParseURL(t, c.right)

			result := SameURL(left, right)

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
