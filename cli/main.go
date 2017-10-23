package main

func main() {
	cli := NewCli().SetFlags()

	// Add all subcommands.
	cli.AddCommand(&PullCommand{})
	cli.AddCommand(&CreateCommand{})
	cli.AddCommand(&StartCommand{})
	cli.AddCommand(&VersionCommand{})
	cli.Run()
}
