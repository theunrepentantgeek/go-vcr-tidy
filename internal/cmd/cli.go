package cmd

type CLI struct {
	Verbose bool  `help:"Enable verbose logging." short:"v"`
	Clean   Clean `cmd:""                         help:"Clean go-vcr cassette files."`
}
