package azure

import (
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

func relinkLocationHeaders(
	interactions []interaction.Interface,
) {
	//nolint:gosec // No risk of slice index out of range; loop condition ensures at least two elements
	for i := range len(interactions) - 1 {
		prior := interactions[i]
		next := interactions[i+1]

		relinkLocationHeader(prior, next)
	}
}

func relinkLocationHeader(
	prior interaction.Interface,
	next interaction.Interface,
) {
	priorURL := prior.Request().FullURL()
	nextURL := next.Request().FullURL()

	if urltool.SameURL(priorURL, nextURL) {
		// Same URL, ensure no Location header present
		prior.Response().RemoveHeader("Location")
	} else {
		// Different URL, ensure Location header present
		prior.Response().SetHeader("Location", nextURL.String())
	}
}
