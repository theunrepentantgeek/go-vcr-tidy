package vcrcleaner

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/rotisserie/eris"
	"go.yaml.in/yaml/v3"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cleaner"
)

// Cleaner is a tool for cleaning go-vcr recordings.
type Cleaner struct {
	core    *cleaner.Cleaner
	mapping map[int]*vcrInteraction
	log     logr.Logger
}

func New(
	log logr.Logger,
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

func (c *Cleaner) CleanFile(path string) error {
	// Remove .yaml from the path if present, as go-vcr expects just the base name
	path = strings.TrimSuffix(path, ".yaml")

	// Load the cassette
	c.log.Info("Cleaning cassette", "path", path)

	cas, err := cassette.Load(path)
	if err != nil {
		return eris.Wrapf(err, "loading cassette from %s", path)
	}

	// Clean the cassette
	modified, err := c.CleanCassette(cas)
	if err != nil {
		return eris.Wrapf(err, "cleaning cassette from %s", path)
	}

	// If modified, save the cassette back
	if modified {
		cas.MarshalFunc = yaml.Marshal // Odd to need to do this explicitly

		err = cas.Save()
		if err != nil {
			return eris.Wrapf(err, "saving cleaned cassette to %s", path)
		}

		c.log.Info("Saved cleaned cassette", "path", path)
	}

	c.log.Info("No change to cassette", "path", path)

	return nil
}

// CleanCassette processes a cassette, marking interactions for removal as needed.
// Returns true if any interactions were marked for removal, false otherwise, along with any error encountered.
func (c *Cleaner) CleanCassette(cas *cassette.Cassette) (bool, error) {
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
	return c.inspect(i)
}

// BeforeSaveHook is the hook to be called before an interaction is saved.
func (c *Cleaner) BeforeSaveHook(i *cassette.Interaction) error {
	c.markIfExcluded(i)

	return nil
}
