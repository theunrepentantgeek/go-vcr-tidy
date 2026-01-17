package vcrcleaner

import (
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/report"
)

// cassetteSummary generates a summary table of the cassette interactions.
// cas is the cassette to summarize.
// columns are additional columns to include in the summary.
func cassetteSummary(
	cas *cassette.Cassette,
	columns ...cassetteColumn,
) string {
	// Build common URL prefix
	prefix := commonURLPrefix(cas)

	headers := []string{"", "Method", "Code", prefix}
	for _, column := range columns {
		headers = append(headers, column.Header)
	}

	// Build a summary of the cassette interactions
	tbl := report.NewMarkdownTable(headers...)

	for _, interaction := range cas.Interactions {
		discarded := ""
		if interaction.DiscardOnSave {
			discarded = "X"
		}

		// Get URL without query parameters and common prefix
		u := strings.TrimPrefix(interaction.Request.URL, prefix)
		if i := strings.Index(u, "?"); i != -1 {
			u = u[:i]
		}

		// Format status code as string
		statusCode := strconv.Itoa(interaction.Response.Code)

		// Write method, URL, and other details, including custom columns (if any)
		row := []string{
			discarded,
			interaction.Request.Method,
			statusCode,
			u,
		}

		for _, column := range columns {
			row = append(row, column.fn(interaction))
		}

		tbl.AddRow(row...)
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

type cassetteColumn struct {
	Header string
	fn     func(*cassette.Interaction) string
}

func withColumn(
	header string,
	fn func(*cassette.Interaction) string,
) cassetteColumn {
	return cassetteColumn{
		Header: header,
		fn:     fn,
	}
}
