package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"

	"github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"
)

type CleanCommand struct {
	Verbose bool `help:"Enable verbose logging."`
	Debug   bool `help:"Enable debug logging."`

	Globs []string        `arg:""   help:"Paths to go-vcr cassette files to clean. Globbing allowed." type:"file"`
	Clean CleaningOptions `embed:"" prefix:"clean-"`
}

// Run executes the clean command for each provided path.
func (c *CleanCommand) Run(ctx *Context) error {
	for _, glob := range c.Globs {
		err := c.cleanFilesByGlob(ctx, glob)
		if err != nil {
			return err
		}
	}

	// Log final summary
	ctx.Log.Info(
		"Cleaning complete",
		"scanned", ctx.FilesScanned,
		"modified", ctx.FilesModified,
	)

	return nil
}

// buildOptions builds the vcrcleaner options based on the CLI flags.
func (c *CleanCommand) buildOptions() ([]vcrcleaner.Option, error) {
	options := c.Clean.Options()

	if len(options) == 0 {
		return nil, eris.New("no cleaning options specified; at least one must be set")
	}

	return options, nil
}

// cleanFilesByGlob cleans any cassette files identified by the given glob path.
func (c *CleanCommand) cleanFilesByGlob(ctx *Context, glob string) error {
	paths, err := filepath.Glob(glob)
	if err != nil {
		return eris.Wrap(err, "failed to glob path")
	}

	// Early exit for no matches
	if len(paths) == 0 {
		ctx.Log.Info("No cassettes found to clean", "glob", glob)

		return nil
	}

	// Log the number of matching files found for globs
	ctx.Log.Info(
		"Found cassettes to clean",
		"count", len(paths),
		"glob", glob)

	// Collect errors, allowing us to attempt processing of all files
	var errs []error

	for _, path := range paths {
		err := c.cleanFile(ctx, path)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return eris.Wrap(
			errors.Join(errs...),
			"one or more errors occurred while cleaning cassette files")
	}

	return nil
}

// cleanFile cleans the cassette file at the specified path.
func (c *CleanCommand) cleanFile(ctx *Context, path string) error {
	options, err := c.buildOptions()
	if err != nil {
		return eris.Wrap(err, "building cleaner options")
	}

	ctx.FilesScanned++

	cleaner := vcrcleaner.New(
		ctx.Log,
		options...,
	)

	modified, err := cleaner.CleanFile(path)
	if err != nil {
		return eris.Wrapf(err, "cleaning cassette file at path %s", path)
	}

	if modified {
		ctx.FilesModified++
	}

	return nil
}

// CreateLogger builds a slog logger configured from the CleanCommand flags.
func (c *CleanCommand) CreateLogger() *slog.Logger {
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	} else if c.Verbose {
		level = vcrcleaner.LevelVerbose
	}

	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewTextHandler(os.Stdout, opts)

	return slog.New(handler)
}
