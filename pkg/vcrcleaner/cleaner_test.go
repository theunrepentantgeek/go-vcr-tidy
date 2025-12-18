package vcrcleaner

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

func TestGolden_CleanerClean_givenRecording_removesExpectedInteractions(t *testing.T) {
	t.Parallel()

	// Analyzers we want to test
	analyzers := map[string]Option{
		"Reduce Delete Monitoring": ReduceDeleteMonitoring(),
	}

	// Find all the *.yaml files under testdata
	// Contains fully qualified path, keyed by filename, without extension
	recordings := map[string]string{}

	files, err := filepath.Glob(filepath.Join("testdata", "*.yaml"))
	if err != nil {
		t.Fatalf("Failed to find test data files: %v", err)
	}

	for _, file := range files {
		baseName := strings.TrimSuffix(filepath.Base(file), ".yaml")
		recordings[baseName] = strings.TrimSuffix(file, ".yaml") // Cassette names do not include .yaml
	}

	// Construct our test matrix - run every option for every file
	type testcase struct {
		option        Option
		recordingPath string
	}

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

	// Run each option as a golden test
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := goldie.New(t, goldie.WithSubTestNameForDir(true))
			log := newTestLogger(t)

			// Load the cassette from the file
			cas, err := cassette.Load(c.recordingPath)
			if err != nil {
				t.Fatalf("Failed to load cassette from %s: %v", c.recordingPath, err)
			}

			// Get baseline YAML for the cassette.
			// We don't use the YAML from the file directly because we don't want our diffs to be polluted by other
			// format changes.
			baseline, err := cassetteToYaml(cas)
			if err != nil {
				t.Fatalf("Failed to get baseline cassette YAML: %v", err)
			}

			// Clean it
			cleaner := New(log, c.option)
			err = cleaner.Clean(cas)
			if err != nil {
				t.Fatalf("Failed to clean cassette: %v", err)
			}

			// Get cleaned YAML for the cassette.
			cleaned, err := cassetteToYaml(cas)
			if err != nil {
				t.Fatalf("Failed to get cleaned cassette YAML: %v", err)
			}

			// Compare it to the original yaml
			d := diff.LineDiff(cleaned, baseline)

			// use goldie to assert the changes made
			g.Assert(t, name, []byte(d))
		})
	}
}
