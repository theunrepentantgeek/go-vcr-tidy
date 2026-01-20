package vcrcleaner

import "net/http"

// vcrResponse represents the response portion of a VCR interaction.
type vcrResponse struct {
	parent *vcrInteraction
}

// StatusCode returns the HTTP status code of the response.
func (r *vcrResponse) StatusCode() int {
	return r.parent.interaction.Response.Code
}

// Header returns the value of the specified response header.
func (r *vcrResponse) Header(name string) (string, bool) {
	headers := r.parent.interaction.Response.Headers
	if headers == nil {
		return "", false
	}

	key := http.CanonicalHeaderKey(name)

	values, ok := headers[key]
	if !ok || len(values) == 0 {
		return "", false
	}

	return values[0], true
}

// SetHeader sets the value of the specified response header.
func (r *vcrResponse) SetHeader(name string, value string) {
	headers := r.parent.interaction.Response.Headers
	if headers == nil {
		headers = make(map[string][]string)
		r.parent.interaction.Response.Headers = headers
	}

	key := http.CanonicalHeaderKey(name)
	headers[key] = []string{value}
}

// RemoveHeader removes the specified response header.
func (r *vcrResponse) RemoveHeader(name string) {
	headers := r.parent.interaction.Response.Headers
	if headers == nil {
		return
	}

	key := http.CanonicalHeaderKey(name)
	delete(headers, key)
}

func (r *vcrResponse) Body() []byte {
	return []byte(r.parent.interaction.Response.Body)
}
