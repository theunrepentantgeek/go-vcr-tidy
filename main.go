package main

import (
	"github.com/alecthomas/kong"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cmd"
)

func main() {
	// Entry point for the application.
	var cli cmd.CleanCommand

	ctx := kong.Parse(&cli,
		kong.UsageOnError())

	cmdCtx := &cmd.Context{
		Verbose: cli.Verbose,
		Debug:   cli.Debug,
		Log:     cli.CreateLogger(),
	}

	err := ctx.Run(cmdCtx)

	ctx.FatalIfErrorf(err)
}
