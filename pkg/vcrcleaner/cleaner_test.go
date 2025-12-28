package vcrcleaner

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/andreyvit/diff"
	"github.com/sebdah/goldie/v2"
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

			// Get baseline summary of the cassette
			baseline := casssetteSummary(cas)

			// Clean it
			cleaner := New(log, c.option)

			err = cleaner.CleanCassette(cas)
			g.Expect(err).NotTo(HaveOccurred(), "cleaning cassette from %s", c.recordingPath)

			// Get cleaned summary for the cassette.
			cleaned := casssetteSummary(cas)

			// Compare it to the original yaml
			d := diff.LineDiff(baseline, cleaned)

			// use goldie to assert the changes made
			gold := goldie.New(t, goldie.WithTestNameForDir(true))
			gold.Assert(t, name, []byte(d))
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
