package vcrcleaner

import "net/url"

// vcrRequest represents the request portion of a VCR interaction.
type vcrRequest struct {
	parent *vcrInteraction
}

// FullURL returns the full URL of the request.
func (r *vcrRequest) FullURL() url.URL {
	result, err := url.Parse(r.parent.interaction.Request.URL)
	if err != nil {
		// If parsing fails, panic (this should never happen in normal operation).
		panic(err)
	}

	return *result
}

// URL returns the base URL of the request (without query parameters or fragment).
func (r *vcrRequest) URL() url.URL {
	result := r.FullURL()
	result.RawQuery = ""
	result.Fragment = ""

	return result
}

func (r *vcrRequest) Method() string {
	return r.parent.interaction.Request.Method
}
