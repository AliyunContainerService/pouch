package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genDocDescription is used to describe gen-doc command in detail and auto generate command doc.
var genDocDescription = "Generate docs for cli command"

// GenDocCommand is used to implement 'gen-doc' command.
type GenDocCommand struct {
	baseCommand
}

// Init initializes GenDocCommand command.
func (g *GenDocCommand) Init(c *Cli) {
	g.cli = c
	g.cmd = &cobra.Command{
		Use:   "gen-doc",
		Short: "Generate docs",
		Long:  genDocDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runGenDoc(args)
		},
		Example: genDocExample(),
	}
	g.addFlags()
}

// addFlags adds flags for specific command.
func (g *GenDocCommand) addFlags() {
}

func (g *GenDocCommand) runGenDoc(args []string) error {
	if _, err := os.Stat("./docs/commandline"); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory docs/commandline does not exist")
		}
		return err
	}
	return doc.GenMarkdownTree(g.cli.rootCmd, "./docs/commandline")
}

// genDocExample shows examples in genDoc command, and is used in auto-generated cli docs.
func genDocExample() string {
	return "$ pouch gen-doc"
}
