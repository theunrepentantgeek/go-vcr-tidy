package vcrcleaner

import (
	"os"

	"github.com/rotisserie/eris"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/yaml.v3"
)

func cassetteToYaml(cas *cassette.Cassette) (string, error) {
	cas.MarshalFunc = yaml.Marshal // Odd to need to do this explicitly

	fs := newMemoryFS()
	err := cas.SaveWithFS(fs)
	if err != nil {
		return "", eris.Wrap(err, "failed to convert cassette to YAML")
	}

	// Get the saved data from the in-memory filesystem
	cleanedBytes, ok := fs.data[cas.Name+".yaml"]
	if !ok {
		return "", eris.New("failed to find saved cassette in memory filesystem")
	}

	return string(cleanedBytes), nil
}

// memoryFS is an in-memory filesystem for testing
type memoryFS struct {
	data map[string][]byte
}

func newMemoryFS() *memoryFS {
	return &memoryFS{
		data: make(map[string][]byte),
	}
}

func (m *memoryFS) ReadFile(name string) ([]byte, error) {
	if data, ok := m.data[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *memoryFS) WriteFile(name string, data []byte) error {
	m.data[name] = data
	return nil
}

func (m *memoryFS) IsFileExists(name string) bool {
	_, ok := m.data[name]
	return ok
}
