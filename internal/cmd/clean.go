package cmd

import (
	"path/filepath"

	"github.com/rotisserie/eris"
	"github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"
)

type Clean struct {
	Globs []string `arg:"" type:"file" help:"Paths to go-vcr cassette files to clean. Globbing allowed."`
	Clean struct {
		Deletes               *bool `help:"Clean delete interactions."`
		LongRunningOperations *bool `help:"Clean Azure long-running operation interactions."`
	} `embed:"" prefix:"clean."`
}

// Run executes the clean command for each provided path.
func (c *Clean) Run(ctx *Context) error {
	for _, glob := range c.Globs {
		err := c.cleanGlob(ctx, glob)
		if err != nil {
			return err
		}
	}

	return nil
}

// buildOptions builds the vcrcleaner options based on the CLI flags.
func (c *Clean) buildOptions() ([]vcrcleaner.Option, error) {
	var options []vcrcleaner.Option

	if c.Clean.Deletes != nil && *c.Clean.Deletes {
		options = append(options, vcrcleaner.ReduceDeleteMonitoring())
	}

	if c.Clean.LongRunningOperations != nil && *c.Clean.LongRunningOperations {
		options = append(options, vcrcleaner.ReduceAzureLongRunningOperationPolling())
	}

	if len(options) == 0 {
		return nil, eris.New("no cleaning options specified; at least one must be set")
	}

	return options, nil
}

// cleanGlob cleans any cassette files identified by the given glob path.
func (c *Clean) cleanGlob(ctx *Context, glob string) error {
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
		err := c.cleanPath(ctx, path)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanPath cleans the cassette file at the specified path.
func (c *Clean) cleanPath(ctx *Context, path string) error {
	options, err := c.buildOptions()
	if err != nil {
		return eris.Wrap(err, "building cleaner options")
	}

	cleaner := vcrcleaner.New(
		ctx.Log.V(1),
		options...,
	)

	ctx.Log.Info("Cleaning cassette", "path", path)

	err = cleaner.CleanFile(path)
	if err != nil {
		return eris.Wrapf(err, "cleaning cassette file at path %s", path)
	}

	ctx.Log.Info("Finished cleaning cassette", "path", path)

	return nil
}
