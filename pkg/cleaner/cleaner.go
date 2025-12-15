package cleaner

import (
	"sync"

	"github.com/google/uuid"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Cleaner is a tool for cleaning go-vcr recordings.
type Cleaner struct {
	// analyzers is a set of active analyzers, keyed by a randomly assigned identifier for tracking.
	analyzers map[uuid.UUID]analyzer.Interface
	// interactionsToRemove is a set of interactions we've selected for removal from the recording
	interactionsToRemove map[uuid.UUID]bool
	// padlock is used to make concurrent access safe
	padlock sync.Mutex
}

// NewCleaner creates a new Cleaner instance with the specified analyzers included.
func NewCleaner(analyzers ...analyzer.Interface) *Cleaner {
	result := &Cleaner{
		analyzers:            make(map[uuid.UUID]analyzer.Interface),
		interactionsToRemove: make(map[uuid.UUID]bool),
	}

	result.Add(analyzers...)

	return result
}

// Add one or more analyzers to the cleaner's active set.
func (c *Cleaner) Add(analyzers ...analyzer.Interface) {
	for _, a := range analyzers {
		// We give each analyzer a unique identifer to make it easy to track them when finished
		c.analyzers[uuid.New()] = a
	}
}

// Remove one or more analyzers from the cleaner's active set.
func (c *Cleaner) remove(ids ...uuid.UUID) {
	for _, id := range ids {
		delete(c.analyzers, id)
	}
}

// exclude adds the specified interations to the set of interactions to be removed.
func (c *Cleaner) exclude(interactions ...interaction.Interface) {
	for _, inter := range interactions {
		c.interactionsToRemove[inter.ID()] = true
	}
}

// analyze processes an interaction through all active analyzers, handling spawning and finishing as needed.
func (c *Cleaner) analyze(interaction interaction.Interface) error {
	var toRemove []uuid.UUID
	var toAdd []analyzer.Interface

	for id, a := range c.analyzers {
		result, err := a.Analyze(interaction)
		if err != nil {
			return err
		}

		if len(result.Excluded) > 0 {
			c.exclude(result.Excluded...)
		}

		if result.Finished {
			toRemove = append(toRemove, id)
		}

		if len(result.Spawn) > 0 {
			toAdd = append(toAdd, result.Spawn...)
		}
	}

	c.remove(toRemove...)
	c.Add(toAdd...)

	return nil
}
