package vcrcleaner

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sebdah/goldie/v2"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

func TestGolden_CleanerClean_givenRecording_removesExpectedInteractions(t *testing.T) {
	t.Parallel()

	// Analyzers we want to test
	analyzers := map[string]Option{
		"Reduce Delete Monitoring":              ReduceDeleteMonitoring(),
		"Reduce Long Running Operation polling": ReduceAzureLongRunningOperationPolling(),
	}

	// Find all the *.yaml files under testdata
	// Contains fully qualified path, keyed by filename, without extension
	recordings := fileTestRecordings(t)

	// Construct our test matrix - run every option for every file
	cases := createTestMatrix(analyzers, recordings)

	// Run each option as a golden test
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			log := newTestLogger(t)

			// Load the cassette from the file
			cas, err := cassette.Load(c.recordingPath)
			g.Expect(err).NotTo(HaveOccurred(), "loading cassette from %s", c.recordingPath)

			// Get baseline YAML for the cassette.
			// We don't use the YAML from the file directly because we don't want our diffs to be polluted by other
			// format changes.
			baseline, err := cassetteToYaml(cas)
			g.Expect(err).NotTo(HaveOccurred(), "getting baseline YAML for cassette from %s", c.recordingPath)

			// Clean it
			cleaner := New(log, c.option)

			err = cleaner.CleanCassette(cas)
			g.Expect(err).NotTo(HaveOccurred(), "cleaning cassette from %s", c.recordingPath)

			// Get cleaned YAML for the cassette.
			cleaned, err := cassetteToYaml(cas)
			g.Expect(err).NotTo(HaveOccurred(), "getting cleaned YAML for cassette from %s", c.recordingPath)

			// Generate a patch diff showing the changes
			patchDiff := generatePatchDiff(baseline, cleaned)

			// use goldie to assert the changes made
			gold := goldie.New(t, goldie.WithTestNameForDir(true))
			gold.Assert(t, name, []byte(patchDiff))
		})
	}
}

func createTestMatrix(analyzers map[string]Option, recordings map[string]string) map[string]testcase {
	cases := map[string]testcase{}

	for analyzerName, option := range analyzers {
		for recordingName, recordingPath := range recordings {
			testName := path.Join(analyzerName, recordingName)
			cases[testName] = testcase{
				option:        option,
				recordingPath: recordingPath,
			}
		}
	}

	return cases
}

func fileTestRecordings(t *testing.T) map[string]string {
	t.Helper()

	recordings := map[string]string{}

	files, err := filepath.Glob(filepath.Join("testdata", "*.yaml"))
	if err != nil {
		t.Fatalf("Failed to find test data files: %v", err)
	}

	for _, file := range files {
		baseName := strings.TrimSuffix(filepath.Base(file), ".yaml")
		recordings[baseName] = strings.TrimSuffix(file, ".yaml") // Cassette names do not include .yaml
	}

	return recordings
}

type testcase struct {
	option        Option
	recordingPath string
}

// generatePatchDiff generates a patch format diff showing changes between baseline and cleaned.
func generatePatchDiff(baseline, cleaned string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(baseline, cleaned, false)
	patches := dmp.PatchMake(baseline, diffs)

	// Convert patches to readable text
	// The patch format uses URL encoding for special characters, so we decode it
	var result strings.Builder

	for _, patch := range patches {
		patchStr := patch.String()
		// Decode URL-encoded characters (like %0A for newlines)
		decoded, err := url.QueryUnescape(patchStr)
		if err != nil {
			// If decoding fails, use the original string
			decoded = patchStr
		}

		result.WriteString(decoded)
	}

	return result.String()
}
