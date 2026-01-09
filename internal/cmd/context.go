package cmd

import "log/slog"

type Context struct {
	Verbose       bool         // Use verbose logging
	Debug         bool         // Use debug logging
	Log           *slog.Logger // Logger to use
	FilesScanned  int          // Number of files scanned
	FilesModified int          // Number of files modified
}
