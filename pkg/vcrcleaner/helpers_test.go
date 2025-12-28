package vcrcleaner

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

func casssetteSummary(cas *cassette.Cassette) string {
	builder := &strings.Builder{}

	for _, interaction := range cas.Interactions {
		// Skip discarded interactions
		if interaction.DiscardOnSave {
			continue
		}

		// Get URL without query parameters
		u := interaction.Request.URL
		if i := strings.Index(u, "?"); i != -1 {
			u = u[:i]
		}

		// Write method and URL
		fmt.Fprintf(builder,
			"%s %d %s\n",
			interaction.Request.Method,
			interaction.Response.Code,
			u)
	}

	return builder.String()
}

// newTestLogger creates a test logger for the given test.
func newTestLogger(t *testing.T) logr.Logger {
	t.Helper()

	return testr.NewWithOptions(t, testr.Options{Verbosity: 1})
}
