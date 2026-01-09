package vcrcleaner

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/rotisserie/eris"
	"go.yaml.in/yaml/v3"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cleaner"
)

// LevelVerbose is a custom log level between INFO and DEBUG.
const LevelVerbose = slog.Level(-2)

// Cleaner is a tool for cleaning go-vcr recordings.
type Cleaner struct {
	core    *cleaner.Cleaner
	mapping map[int]*vcrInteraction
	log     *slog.Logger
	padlock sync.Mutex
}

func New(
	log *slog.Logger,
	options ...Option,
) *Cleaner {
	result := &Cleaner{
		core:    cleaner.New(),
		mapping: make(map[int]*vcrInteraction),
		log:     log,
	}

	for _, option := range options {
		option(result.core)
	}

	return result
}

// CleanFile processes a single cassette file, removing unnecessary interactions.
// The path parameter should be the full path to the cassette file, including the .yaml extension.
// Returns true if the file was modified and saved, false if no changes were made.
// Returns an error if the file cannot be processed.
func (c *Cleaner) CleanFile(
	path string,
) (bool, error) {
	// Remove .yaml from the path if present, as go-vcr expects just the base name
	cassetteName := strings.TrimSuffix(path, ".yaml")

	// Attempt to load a cassette from the specified path
	// This might fail if we are given a different kind of YAML file, so we need to handle that gracefully
	cas, err := cassette.Load(cassetteName)
	if err != nil {
		c.log.Warn("Skipping non-cassette file", "path", path, "error", err)

		return false, nil
	}

	c.log.Log(context.Background(), LevelVerbose, "Cleaning cassette", "path", path)

	// Clean the cassette
	modified, err := c.CleanCassette(cas)
	if err != nil {
		return false, eris.Wrapf(err, "cleaning cassette from %s", path)
	}

	// If modified, save the cassette back
	if modified {
		cas.MarshalFunc = yaml.Marshal // Odd to need to do this explicitly

		err = cas.Save()
		if err != nil {
			return false, eris.Wrapf(err, "saving cleaned cassette to %s", path)
		}

		c.log.Info("Saved cleaned cassette", "path", path)
	} else {
		c.log.Log(context.Background(), LevelVerbose, "No change to cassette", "path", path)
	}

	return modified, nil
}

// CleanCassette processes a cassette, marking interactions for removal as needed.
// Returns true if any interactions were marked for removal, false otherwise, along with any error encountered.
func (c *Cleaner) CleanCassette(cas *cassette.Cassette) (bool, error) {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	// Scan all interactions
	for _, i := range cas.Interactions {
		if err := c.inspect(i); err != nil {
			return false, eris.Wrapf(err, "inspecting interaction %d", i.ID)
		}
	}

	// If any interactions are to be marked for removal, mark them now
	if c.core.InteractionsToRemove() > 0 {
		for _, i := range cas.Interactions {
			c.markIfExcluded(i)
		}

		return true, nil
	}

	// No interactions were marked for removal
	return false, nil
}

// inspect processes a single interaction through the cleaner.
func (c *Cleaner) inspect(i *cassette.Interaction) error {
	vi := newVCRInteraction(i)
	c.mapping[i.ID] = vi

	err := c.core.Analyze(c.log, vi)
	if err != nil {
		return eris.Wrapf(err, "analyzing interaction ID %d", i.ID)
	}

	return nil
}

// markIfExcluded marks an interaction for removal, if needed.
func (c *Cleaner) markIfExcluded(i *cassette.Interaction) {
	vi, ok := c.mapping[i.ID]
	if !ok {
		// Not an interaction we know about; nothing to do.
		return
	}

	if c.core.ShouldRemove(vi) {
		i.DiscardOnSave = true
	}
}

// AfterCaptureHook is the hook to be called after an interaction is captured.
func (c *Cleaner) AfterCaptureHook(i *cassette.Interaction) error {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	return c.inspect(i)
}

// BeforeSaveHook is the hook to be called before an interaction is saved.
func (c *Cleaner) BeforeSaveHook(i *cassette.Interaction) error {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	c.markIfExcluded(i)

	return nil
}
