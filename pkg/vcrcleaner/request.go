package vcrcleaner

import (
	"net/url"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/urltool"
)

// vcrRequest represents the request portion of a VCR interaction.
type vcrRequest struct {
	parent *vcrInteraction
}

// FullURL returns the full URL of the request.
func (r *vcrRequest) FullURL() *url.URL {
	result, err := url.Parse(r.parent.interaction.Request.URL)
	if err != nil {
		// If parsing fails, panic (this should never happen in normal operation).
		panic(err)
	}

	return result
}

// BaseURL returns the base BaseURL of the request (without query parameters or fragment).
func (r *vcrRequest) BaseURL() *url.URL {
	return urltool.BaseURL(r.FullURL())
}

func (r *vcrRequest) Method() string {
	return r.parent.interaction.Request.Method
}
