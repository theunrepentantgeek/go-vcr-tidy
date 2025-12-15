package cleaner

import (
	"uuid"

	"github.com/google/uuid"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// Cleaner is a tool for cleaning go-vcr recordings.
type Cleaner struct {
	// analyzers is a set of active analyzers, keyed by a randomly assigned identifier for tracking.
	analyzers            map[string]analyzer.Interface
	interactionsToRemove map[interaction.Interface]bool
}

// NewCleaner creates a new Cleaner instance with the specified analyzers included.
func NewCleaner(analyzers ...analyzer.Interface) *Cleaner {
	result := &Cleaner{
		analyzers:            make(map[string]analyzer.Interface),
		interactionsToRemove: make(map[interaction.Interface]bool),
	}

	result.Add(analyzers...)

	return result
}

// Add one or more analyzers to the cleaner's active set.
func (c *Cleaner) Add(analyzers ...analyzer.Interface) {
	for _, a := range analyzers {
		id := uuid.NewString()
		c.analyzers[id] = a
	}
}

// Remove one or more analyzers from the cleaner's active set.
func (c *Cleaner) remove(ids ...string) {
	for _, id := range ids {
		delete(c.analyzers, id)
	}
}

// exclude adds the specified interations to the set of interactions to be removed.
func (c *Cleaner) exclude(interactions ...interaction.Interface) {
	for _, inter := range interactions {
		c.interactionsToRemove[inter] = true
	}
}

// analyze processes an interaction through all active analyzers, handling spawning and finishing as needed.
func (c *Cleaner) analyze(interaction interaction.Interface) error {
	var toRemove []string
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
