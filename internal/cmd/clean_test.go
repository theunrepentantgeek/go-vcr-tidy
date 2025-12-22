package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr/testr"
)

// buildOptions Tests

func TestBuildOptions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deletes                *bool
		longRunningOperations  *bool
		expectError            bool
		expectedOptionsCount   int
		expectedErrorSubstring string
	}{
		"WithNoOptionsSpecified_ReturnsError": {
			deletes:                nil,
			longRunningOperations:  nil,
			expectError:            true,
			expectedOptionsCount:   0,
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithOnlyDeletesSet_ReturnsDeleteOption": {
			deletes:              boolPtr(true),
			longRunningOperations: nil,
			expectError:          false,
			expectedOptionsCount: 1,
		},
		"WithOnlyLongRunningOperationsSet_ReturnsLROOption": {
			deletes:               nil,
			longRunningOperations: boolPtr(true),
			expectError:           false,
			expectedOptionsCount:  1,
		},
		"WithBothOptionsSet_ReturnsBothOptions": {
			deletes:               boolPtr(true),
			longRunningOperations: boolPtr(true),
			expectError:           false,
			expectedOptionsCount:  2,
		},
		"WithDeletesSetToFalse_ReturnsError": {
			deletes:                boolPtr(false),
			longRunningOperations:  nil,
			expectError:            true,
			expectedOptionsCount:   0,
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithLongRunningOperationsSetToFalse_ReturnsError": {
			deletes:                nil,
			longRunningOperations:  boolPtr(false),
			expectError:            true,
			expectedOptionsCount:   0,
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

			if c.expectError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(c.expectedErrorSubstring))
				g.Expect(options).To(BeNil())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(options).To(HaveLen(c.expectedOptionsCount))
			}
		})
	}
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}

// cleanGlob Tests

func TestCleanGlob_WithNoMatchingFiles_LogsAndSucceeds(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	// Use a glob pattern that won't match anything
	glob := filepath.Join(tmpDir, "nonexistent-*.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithSingleMatchingFile_CleansFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test cassette file
	cassetteContent := `---
version: 2
interactions: []
`
	cassettePath := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
	g.Expect(err).ToNot(HaveOccurred())

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	// Use a glob pattern that matches the file
	glob := filepath.Join(tmpDir, "test.yaml")
	err = c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())

	// Verify the file still exists (cleaned in place)
	_, err = os.Stat(cassettePath)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithMultipleMatchingFiles_CleansAllFiles(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create multiple test cassette files
	cassetteContent := `---
version: 2
interactions: []
`
	for i := 1; i <= 3; i++ {
		cassettePath := filepath.Join(tmpDir, "test"+strconv.Itoa(i)+".yaml")
		err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
		g.Expect(err).ToNot(HaveOccurred())
	}

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	// Use a glob pattern that matches all files
	glob := filepath.Join(tmpDir, "test*.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanGlob_WithInvalidGlobPattern_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	// Use an invalid glob pattern (malformed character class)
	glob := filepath.Join(t.TempDir(), "test[.yaml")
	err := c.cleanGlob(ctx, glob)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("failed to glob path"))
}

// cleanPath Tests

func TestCleanPath_WithValidCassette_CleansSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test cassette file
	cassetteContent := `---
version: 2
interactions: []
`
	cassettePath := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
	g.Expect(err).ToNot(HaveOccurred())

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err = c.cleanPath(ctx, cassettePath)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestCleanPath_WithNoOptionsSet_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test cassette file
	cassetteContent := `---
version: 2
interactions: []
`
	cassettePath := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
	g.Expect(err).ToNot(HaveOccurred())

	c := &Clean{}

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err = c.cleanPath(ctx, cassettePath)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("building cleaner options"))
}

func TestCleanPath_WithNonexistentFile_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.cleanPath(ctx, filepath.Join(t.TempDir(), "nonexistent", "path", "cassette.yaml"))

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("cleaning cassette file"))
}

// Run Tests

func TestRun_WithSingleGlob_ProcessesSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test cassette file
	cassetteContent := `---
version: 2
interactions: []
`
	cassettePath := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
	g.Expect(err).ToNot(HaveOccurred())

	c := &Clean{
		Globs: []string{cassettePath},
	}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err = c.Run(ctx)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun_WithMultipleGlobs_ProcessesAllSuccessfully(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create multiple test cassette files
	cassetteContent := `---
version: 2
interactions: []
`
	var globs []string
	for i := 1; i <= 3; i++ {
		cassettePath := filepath.Join(tmpDir, "test"+strconv.Itoa(i)+".yaml")
		err := os.WriteFile(cassettePath, []byte(cassetteContent), 0600)
		g.Expect(err).ToNot(HaveOccurred())
		globs = append(globs, cassettePath)
	}

	c := &Clean{
		Globs: globs,
	}
	deletes := true
	c.Clean.Deletes = &deletes

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
		Globs: []string{filepath.Join(t.TempDir(), "test[.yaml")}, // Invalid glob pattern
	}
	deletes := true
	c.Clean.Deletes = &deletes

	ctx := &Context{
		Log: testr.NewWithOptions(t, testr.Options{Verbosity: 1}),
	}

	err := c.Run(ctx)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("failed to glob path"))
}
