package main

import (
	"log/slog"
	"os"
)

// LevelVerbose is a custom log level between INFO and DEBUG
// It is used to log additional information that is more detailed than INFO
// but not as detailed as DEBUG.
const LevelVerbose = slog.Level(-2)

// CreateLogger creates a standard slog logger.
//
//nolint:revive // temporary suppression
func CreateLogger(verbose bool, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	} else if verbose {
		level = LevelVerbose
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	return slog.New(handler)
}
