package main

import (
	"log/slog"
	"os"

	"github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"
)

// CreateLogger creates a standard slog logger.
//
//nolint:revive // temporary suppression
func CreateLogger(verbose bool, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	} else if verbose {
		level = vcrcleaner.LevelVerbose
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	return slog.New(handler)
}
