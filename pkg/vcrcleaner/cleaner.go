package vcrcleaner

import (
	"github.com/go-logr/logr"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cleaner"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
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

func (c *Cleaner) Clean(cas *cassette.Cassette) error {
	for _, i := range cas.Interactions {
		if err := c.inspect(i); err != nil {
			return err
		}
	}

	for _, i := range cas.Interactions {
		c.markIfExcluded(i)
	}

	return nil
}

// inspect processes a single interaction through the cleaner.
func (c *Cleaner) inspect(i *cassette.Interaction) error {
	vi := newVCRInteraction(i)
	c.mapping[i.ID] = vi
	return c.core.Analyze(c.log, vi)
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

func (c *Cleaner) AfterCaptureHook(i *cassette.Interaction) error {
	return c.inspect(i)
}

func (c *Cleaner) BeforeSaveHook(i *cassette.Interaction) error {
	c.markIfExcluded(i)
	return nil
}
