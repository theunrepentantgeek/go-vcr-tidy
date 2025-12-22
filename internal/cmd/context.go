package cmd

import "github.com/go-logr/logr"

type Context struct {
	Verbose bool        // Use verbose logging
	Log     logr.Logger // Logger to use
}
