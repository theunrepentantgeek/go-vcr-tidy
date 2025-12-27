package cmd

type CLI struct {
	Verbose bool         `help:"Enable verbose logging." short:"v"`
	Clean   CleanCommand `cmd:""                         help:"Clean go-vcr cassette files."`
}
