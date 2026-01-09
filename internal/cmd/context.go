package cmd

import "log/slog"

type Context struct {
	Verbose bool         // Use verbose logging
	Log     *slog.Logger // Logger to use
}
