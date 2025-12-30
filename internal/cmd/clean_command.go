package cmd

import (
	"path/filepath"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"
)

type CleanCommand struct {
	Globs []string        `arg:""   help:"Paths to go-vcr cassette files to clean. Globbing allowed." type:"file"`
	Clean CleaningOptions `embed:"" prefix:"clean."`
}

type CleaningOptions struct {
	Deletes               *bool `help:"Clean delete interactions."`
	LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
	ResourceModifications *bool `help:"Clean Azure resource modification (PUT/PATCH) monitoring interactions."`
	ResourceDeletions     *bool `help:"Clean Azure resource deletion monitoring interactions."`
}

// Run executes the clean command for each provided path.
func (c *CleanCommand) Run(ctx *Context) error {
	for _, glob := range c.Globs {
		err := c.cleanFilesByGlob(ctx, glob)
		if err != nil {
			return err
		}
	}

	return nil
}

// buildOptions builds the vcrcleaner options based on the CLI flags.
func (c *CleanCommand) buildOptions() ([]vcrcleaner.Option, error) {
	options := c.collectOptions()

	if len(options) == 0 {
		return nil, eris.New("no cleaning options specified; at least one must be set")
	}

	return options, nil
}

// collectOptions collects all enabled options from the CLI flags.
//
//nolint:revive // Simple sequential flag checks are acceptable
func (c *CleanCommand) collectOptions() []vcrcleaner.Option {
	var options []vcrcleaner.Option

	if c.Clean.Deletes != nil && *c.Clean.Deletes {
		options = append(options, vcrcleaner.ReduceDeleteMonitoring())
	}

	if c.Clean.LongRunningOperations != nil && *c.Clean.LongRunningOperations {
		options = append(options, vcrcleaner.ReduceAzureLongRunningOperationPolling())
	}

	if c.Clean.ResourceModifications != nil && *c.Clean.ResourceModifications {
		options = append(options, vcrcleaner.ReduceAzureResourceModificationMonitoring())
	}

	if c.Clean.ResourceDeletions != nil && *c.Clean.ResourceDeletions {
		options = append(options, vcrcleaner.ReduceAzureResourceDeletionMonitoring())
	}

	return options
}

// cleanFilesByGlob cleans any cassette files identified by the given glob path.
func (c *CleanCommand) cleanFilesByGlob(ctx *Context, glob string) error {
	paths, err := filepath.Glob(glob)
	if err != nil {
		return eris.Wrap(err, "failed to glob path")
	}

	// Early exit for no matches
	if len(paths) == 0 {
		ctx.Log.V(1).Info("No cassettes found to clean", "glob", glob)

		return nil
	}

	if len(paths) > 1 {
		// Multiple matches, log details
		ctx.Log.V(1).Info(
			"Found cassettes to clean",
			"count", len(paths),
			"glob", glob)
	}

	for _, path := range paths {
		err := c.cleanFile(ctx, path)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanFile cleans the cassette file at the specified path.
func (c *CleanCommand) cleanFile(ctx *Context, path string) error {
	options, err := c.buildOptions()
	if err != nil {
		return eris.Wrap(err, "building cleaner options")
	}

	cleaner := vcrcleaner.New(
		ctx.Log.V(1),
		options...,
	)

	err = cleaner.CleanFile(path)
	if err != nil {
		return eris.Wrapf(err, "cleaning cassette file at path %s", path)
	}

	return nil
}
