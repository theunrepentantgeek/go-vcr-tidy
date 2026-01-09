package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/neilotoole/slogt"
	. "github.com/onsi/gomega"
)

/*
 * Helper functions for testing
 */

// newTestLogger creates a test logger for the given test.
func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()

	return slogt.New(t)
}

// buildOptions Tests

//nolint:funlen // Table test cases are extensive but clear
func TestBuildOptions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deletes                *bool
		longRunningOperations  *bool
		resourceModifications  *bool
		resourceDeletions      *bool
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
		"WithOnlyResourceModificationsSet_ReturnsResourceModificationOption": {
			resourceModifications: toPtr(true),
			expectedOptionsCount:  1,
		},
		"WithOnlyResourceDeletionsSet_ReturnsResourceDeletionOption": {
			resourceDeletions:    toPtr(true),
			expectedOptionsCount: 1,
		},
		"WithBothOptionsSet_ReturnsBothOptions": {
			deletes:               toPtr(true),
			longRunningOperations: toPtr(true),
			expectedOptionsCount:  2,
		},
		"WithAllOptionsSet_ReturnsAllOptions": {
			deletes:               toPtr(true),
			longRunningOperations: toPtr(true),
			resourceModifications: toPtr(true),
			resourceDeletions:     toPtr(true),
			expectedOptionsCount:  4,
		},
		"WithDeletesSetToFalse_ReturnsError": {
			deletes:                toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithLongRunningOperationsSetToFalse_ReturnsError": {
			longRunningOperations:  toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithResourceModificationsSetToFalse_ReturnsError": {
			resourceModifications:  toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithResourceDeletionsSetToFalse_ReturnsError": {
			resourceDeletions:      toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			cmd := &CleanCommand{}
			cmd.Clean.Deletes = c.deletes
			cmd.Clean.Azure.LongRunningOperations = c.longRunningOperations
			cmd.Clean.Azure.ResourceModifications = c.resourceModifications
			cmd.Clean.Azure.ResourceDeletions = c.resourceDeletions

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
	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	glob := filepath.Join(tmpDir, "nonexistent-*.yaml")
	err := c.cleanFilesByGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithSingleMatchingFile_CleansFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")
	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	glob := filepath.Join(tmpDir, "test.yaml")
	err := c.cleanFilesByGlob(ctx, glob)

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

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	glob := filepath.Join(tmpDir, "test*.yaml")
	err := c.cleanFilesByGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithInvalidGlobPattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	glob := filepath.Join(t.TempDir(), "test[.yaml")
	err := c.cleanFilesByGlob(ctx, glob)

	g.Expect(err).To(MatchError(ContainSubstring("failed to glob path")))
}

// cleanPath Tests

func TestCleanPath_WithValidCassette_CleansSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")
	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.cleanFile(ctx, cassettePath)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanPath_WithNoOptionsSet_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &CleanCommand{}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.cleanFile(ctx, cassettePath)

	g.Expect(err).To(MatchError(ContainSubstring("building cleaner options")))
}

func TestCleanPath_WithNonexistentFile_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.cleanFile(ctx, filepath.Join(t.TempDir(), "nonexistent", "path", "cassette.yaml"))

	g.Expect(err).To(MatchError(ContainSubstring("cleaning cassette file")))
}

// Run Tests

func TestRun_WithSingleGlob_ProcessesSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	cassettePath := createTestRecording(t, g, tmpDir, "test.yaml")

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
		Globs: []string{cassettePath},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithMultipleGlobs_ProcessesAllSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()

	globs := make([]string, 0, 3)

	for i := 1; i <= 3; i++ {
		cassettePath := createTestRecording(t, g, tmpDir, "test"+strconv.Itoa(i)+".yaml")
		globs = append(globs, cassettePath)
	}

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
		Globs: globs,
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithNoGlobs_SucceedsWithoutProcessing(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &CleanCommand{
		Globs: []string{},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithErrorInGlob_PropagatesError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &CleanCommand{
		Clean: CleaningOptions{
			Deletes: toPtr(true),
		},
		Globs: []string{filepath.Join(t.TempDir(), "test[.yaml")},
	}

	ctx := &Context{
		Log: newTestLogger(t),
	}

	err := c.Run(ctx)

	g.Expect(err).To(MatchError(ContainSubstring("failed to glob path")))
}
