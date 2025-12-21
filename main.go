package main

import (
	"github.com/alecthomas/kong"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cmd"
)

func main() {
	// Entry point for the application.
	var cli cmd.CLI
	ctx := kong.Parse(&cli)
	err := ctx.Run(&cmd.Context{
		Verbose: cli.Verbose,
		Log:     CreateLogger(cli.Verbose),
	})

	ctx.FatalIfErrorf(err)
}
