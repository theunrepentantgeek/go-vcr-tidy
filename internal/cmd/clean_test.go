package cmd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr/testr"
)

// buildOptions Tests

func TestBuildOptions_WithNoOptionsSpecified_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := &Clean{}

	options, err := c.buildOptions()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("no cleaning options specified"))
	g.Expect(options).To(BeNil())
}

func TestBuildOptions_WithOnlyDeletesSet_ReturnsDeleteOption(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	deletes := true
	c := &Clean{}
	c.Clean.Deletes = &deletes

	options, err := c.buildOptions()

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(options).To(HaveLen(1))
}

func TestBuildOptions_WithOnlyLongRunningOperationsSet_ReturnsLROOption(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	lro := true
	c := &Clean{}
	c.Clean.LongRunningOperations = &lro

	options, err := c.buildOptions()

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(options).To(HaveLen(1))
}

func TestBuildOptions_WithBothOptionsSet_ReturnsBothOptions(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	deletes := true
	lro := true
	c := &Clean{}
	c.Clean.Deletes = &deletes
	c.Clean.LongRunningOperations = &lro

	options, err := c.buildOptions()

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(options).To(HaveLen(2))
}

func TestBuildOptions_WithDeletesSetToFalse_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	deletes := false
	c := &Clean{}
	c.Clean.Deletes = &deletes

	options, err := c.buildOptions()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("no cleaning options specified"))
	g.Expect(options).To(BeNil())
}

func TestBuildOptions_WithLongRunningOperationsSetToFalse_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	lro := false
	c := &Clean{}
	c.Clean.LongRunningOperations = &lro

	options, err := c.buildOptions()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("no cleaning options specified"))
	g.Expect(options).To(BeNil())
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
		cassettePath := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".yaml")
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
	glob := "/tmp/test[.yaml"
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

	err := c.cleanPath(ctx, "/nonexistent/path/to/cassette.yaml")

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
		cassettePath := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".yaml")
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
		Globs: []string{"/tmp/test[.yaml"}, // Invalid glob pattern
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
