package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr/testr"
)

// toPtr returns a pointer to the given value.
func toPtr[T any](v T) *T {
	return &v
}

// createTestRecording creates a sample cassette file in the given directory.
func createTestRecording(t *testing.T, g Gomega, tmpDir, filename string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", "sample.yaml"))
	g.Expect(err).ToNot(HaveOccurred())

	cassettePath := filepath.Join(tmpDir, filename)
	err = os.WriteFile(cassettePath, content, 0600)
	g.Expect(err).ToNot(HaveOccurred())

	return cassettePath
}

// buildOptions Tests

func TestBuildOptions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deletes                *bool
		longRunningOperations  *bool
		expectedOptionsCount   int
		expectedErrorSubstring string
	}{
		"WithNoOptionsSpecified_ReturnsError": {
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithOnlyDeletesSet_ReturnsDeleteOption": {
			deletes:              toPtr(true),
			expectedOptionsCount: 1,
		},
		"WithOnlyLongRunningOperationsSet_ReturnsLROOption": {
			longRunningOperations: toPtr(true),
			expectedOptionsCount:  1,
		},
		"WithBothOptionsSet_ReturnsBothOptions": {
			deletes:               toPtr(true),
			longRunningOperations: toPtr(true),
			expectedOptionsCount:  2,
		},
		"WithDeletesSetToFalse_ReturnsError": {
			deletes:                toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithLongRunningOperationsSetToFalse_ReturnsError": {
			longRunningOperations:  toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			cmd := &Clean{}
			cmd.Clean.Deletes = c.deletes
			cmd.Clean.LongRunningOperations = c.longRunningOperations

			options, err := cmd.buildOptions()

			if c.expectedErrorSubstring != "" {
				g.Expect(err).To(MatchError(ContainSubstring(c.expectedErrorSubstring)))
				g.Expect(options).To(BeNil())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(options).To(HaveLen(c.expectedOptionsCount))
			}
		})
	}
}

// cleanGlob Tests

func TestCleanGlob_WithNoMatchingFiles_LogsAndSucceeds(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	glob := filepath.Join(tmpDir, "nonexistent-*.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithSingleMatchingFile_CleansFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	glob := filepath.Join(tmpDir, "test.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())

	// Verify the file still exists (cleaned in place)
	_, err = os.Stat(cassettePath)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithMultipleMatchingFiles_CleansAllFiles(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()

	for i := 1; i <= 3; i++ {
		createTestRecording(t, g, tmpDir, "test"+strconv.Itoa(i)+".yaml")
	}

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	glob := filepath.Join(tmpDir, "test*.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithInvalidGlobPattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	glob := filepath.Join(t.TempDir(), "test[.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).To(MatchError(ContainSubstring("failed to glob path")))
}

// cleanPath Tests

func TestCleanPath_WithValidCassette_CleansSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.cleanPath(ctx, cassettePath)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanPath_WithNoOptionsSet_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &Clean{}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.cleanPath(ctx, cassettePath)

	g.Expect(err).To(MatchError(ContainSubstring("building cleaner options")))
}

func TestCleanPath_WithNonexistentFile_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.cleanPath(ctx, filepath.Join(t.TempDir(), "nonexistent", "path", "cassette.yaml"))

	g.Expect(err).To(MatchError(ContainSubstring("cleaning cassette file")))
}

// Run Tests

func TestRun_WithSingleGlob_ProcessesSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &Clean{
		Globs: []string{cassettePath},
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithMultipleGlobs_ProcessesAllSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()

	var globs []string
	for i := 1; i <= 3; i++ {
		cassettePath := createTestRecording(t, g, tmpDir, "test"+strconv.Itoa(i)+".yaml")
		globs = append(globs, cassettePath)
	}

	c := &Clean{
		Globs: globs,
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithNoGlobs_SucceedsWithoutProcessing(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{
		Globs: []string{},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithErrorInGlob_PropagatesError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{
		Globs: []string{filepath.Join(t.TempDir(), "test[.yaml")},
		Clean: struct {
			Deletes               *bool `help:"Clean delete interactions."`
			LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
		}{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.Run(ctx)

	g.Expect(err).To(MatchError(ContainSubstring("failed to glob path")))
}
