package main

import (
	"github.com/alecthomas/kong"
	"github.com/go-logr/zerologr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cmd"
)

func main() {
	// Entry point for the application.
	var cli cmd.CLI

	ctx := kong.Parse(&cli)

	err := ctx.Run(&cmd.Context{
		Verbose: cli.Verbose,
		Log:     CreateLogger(),
	})

	if cli.Verbose {
		zerologr.SetMaxV(1)
	}

	ctx.FatalIfErrorf(err)
}
