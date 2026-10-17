package cmd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
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

	root, err := os.OpenRoot(tmpDir)
	g.Expect(err).ToNot(HaveOccurred())
	t.Cleanup(func() {
		g.Expect(root.Close()).To(Succeed())
	})

	err = root.WriteFile(filename, content, 0o600)
	g.Expect(err).ToNot(HaveOccurred())

	return filepath.Join(tmpDir, filename)
}
