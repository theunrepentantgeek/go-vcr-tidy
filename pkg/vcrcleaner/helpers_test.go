package vcrcleaner

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

func casssetteSummary(cas *cassette.Cassette) string {
	// Build a summary of the cassette interactions
	builder := &strings.Builder{}

	prefix := commonURLPrefix(cas)
	builder.WriteString(
		fmt.Sprintf(
			"Common URL prefix: %s\n\n",
			prefix))

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

		// Remove common prefix
		u = strings.TrimPrefix(u, prefix)

		// Write method and URL
		fmt.Fprintf(builder,
			"%s %d %s\n",
			interaction.Request.Method,
			interaction.Response.Code,
			u)
	}

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
func newTestLogger(t *testing.T) logr.Logger {
	t.Helper()

	return testr.NewWithOptions(t, testr.Options{Verbosity: 1})
}

// DiffPrettyText converts a []Diff into a text report.
func diffsToText(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			writeDiff(&buff, diff.Text, "+")
		case diffmatchpatch.DiffDelete:
			writeDiff(&buff, diff.Text, "-")
		case diffmatchpatch.DiffEqual:
			writeDiff(&buff, diff.Text, " ")
		}
	}

	return buff.String()
}

func writeDiff(
	buffer *bytes.Buffer,
	text string,
	prefix string,
) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if len(strings.TrimSpace(prefix)) > 0 || len(line) > 0 {
			_, _ = buffer.WriteString(prefix)
			_, _ = buffer.WriteString(line)
		}

		if i < len(lines)-1 {
			_, _ = buffer.WriteString("\n")
		}
	}
}
