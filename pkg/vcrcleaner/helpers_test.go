package vcrcleaner

import (
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/report"
)

func cassetteSummary(cas *cassette.Cassette) string {
	// Build common URL prefix
	prefix := commonURLPrefix(cas)

	// Build a summary of the cassette interactions
	tbl := report.NewMarkdownTable(
		"",
		"Method",
		"Code",
		prefix)

	for _, interaction := range cas.Interactions {
		discard := ""
		if interaction.DiscardOnSave {
			discard = "X"
		}

		// Get URL without query parameters and common prefix
		u := strings.TrimPrefix(interaction.Request.URL, prefix)
		if i := strings.Index(u, "?"); i != -1 {
			u = u[:i]
		}

		// Format status code as string
		statusCode := strconv.Itoa(interaction.Response.Code)

		// Write method and URL
		tbl.AddRow(
			discard,
			interaction.Request.Method,
			statusCode,
			u)
	}

	var builder strings.Builder
	tbl.WriteTo(&builder)

	return builder.String()
}

// commonURLPrefix returns the common URL prefix for all interactions in the cassette.
// This is used to reduce the level of "noise" in golden files.
func commonURLPrefix(cas *cassette.Cassette) string {
	if len(cas.Interactions) == 0 {
		return ""
	}

	dmp := diffmatchpatch.New()

	prefix := cas.Interactions[0].Request.URL
	for _, interaction := range cas.Interactions[1:] {
		u := interaction.Request.URL
		l := dmp.DiffCommonPrefix(prefix, u)
		prefix = prefix[:l]
	}

	return prefix
}

// newTestLogger creates a test logger for the given test.
func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()

	return slogt.New(t)
}
