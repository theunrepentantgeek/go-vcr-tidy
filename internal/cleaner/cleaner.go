package cleaner

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"

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

// New creates a new Cleaner instance with the specified analyzers included.
func New(analyzers ...analyzer.Interface) *Cleaner {
	result := &Cleaner{
		analyzers:            make(map[uuid.UUID]analyzer.Interface),
		interactionsToRemove: make(map[uuid.UUID]bool),
	}

	result.AddAnalyzers(analyzers...)

	return result
}

// AddAnalyzers one or more analyzers to the cleaner's active set.
func (c *Cleaner) AddAnalyzers(analyzers ...analyzer.Interface) {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	for _, a := range analyzers {
		// We give each analyzer a unique identifier to make it easy to track them when finished
		c.analyzers[uuid.New()] = a
	}
}

// Remove one or more analyzers from the cleaner's active set.
func (c *Cleaner) remove(ids ...uuid.UUID) {
	for _, id := range ids {
		delete(c.analyzers, id)
	}
}

// exclude adds the specified interactions to the set of interactions to be removed.
func (c *Cleaner) exclude(interactions ...interaction.Interface) {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	for _, inter := range interactions {
		c.interactionsToRemove[inter.ID()] = true
	}
}

// Analyze processes an interaction through all active analyzers, handling spawning and finishing as needed.
func (c *Cleaner) Analyze(
	log logr.Logger,
	i interaction.Interface,
) error {
	var (
		toRemove []uuid.UUID
		toAdd    []analyzer.Interface
	)

	for id, a := range c.analyzers {
		result, err := a.Analyze(log, i)
		if err != nil {
			return eris.Wrapf(err, "analyzing interaction ID %s", i.ID())
		}

		if result.Finished {
			toRemove = append(toRemove, id)
		}

		// Add any spawned analyzers (if any)
		toAdd = append(toAdd, result.Spawn...)

		// Exclude any interactions marked for exclusion (if any)
		c.exclude(result.Excluded...)
	}

	c.remove(toRemove...)
	c.Add(toAdd...)

	return nil
}

func (c *Cleaner) ShouldRemove(i interaction.Interface) bool {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	_, ok := c.interactionsToRemove[i.ID()]

	return ok
}

// InteractionsToRemove returns the number of interactions marked for removal.
func (c *Cleaner) InteractionsToRemove() int {
	c.padlock.Lock()
	defer c.padlock.Unlock()

	return len(c.interactionsToRemove)
}
