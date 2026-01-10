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

	log := cli.CreateLogger()

	cmdCtx := &cmd.Context{
		Verbose: cli.Verbose,
		Debug:   cli.Debug,
		Log:     log,
	}

	err := ctx.Run(cmdCtx)
	if err != nil {
		cmdCtx.Log.Error("Error executing command", "error", err)
		ctx.Exit(1)
	}

	ctx.Exit(0)
}
