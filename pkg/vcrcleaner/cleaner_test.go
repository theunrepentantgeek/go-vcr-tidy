package vcrcleaner

import (
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

	// Construct our test matrix - run every option for every file
	cases := map[string]struct {
		option        Option
		recordingPath string
	}{}

	// Run each option as a golden test
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := goldie.New(t, goldie.WithSubTestNameForDir(true))

			// Create a cassette from it
			cassette.Load()

			// Clean it
			cleaner := New(c.option)

			// Serialize it back to YAML
			cleaned := "" // serialized cleaned cassette

			// Load the original recording
			yamls := "" // load from c.recordingPath

			// Compare it to the original yaml

			diff.LineDiff(yaml, cleaned)

			// use goldie to assert the result
			g.Assert(t, name, []byte("resulting yaml content"))
		})
	}

}
