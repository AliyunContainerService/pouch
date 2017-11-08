package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// PsCommand is used to implement 'ps' command.
type PsCommand struct {
	baseCommand
}

// Init initializes PsCommand command.
func (p *PsCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "ps",
		Short: "list all containers",
	}
}

// Run is the entry of PsCommand command.
func (p *PsCommand) Run(args []string) {
	apiClient := p.cli.Client()

	containers, err := apiClient.ContainerList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get container list: %v\n", err)
		return
	}

	display := p.cli.NewTableDisplay()
	display.AddRow([]string{"Name", "ID", "Status", "Image"})
	for _, c := range containers {
		display.AddRow([]string{c.Names[0], c.ID[:6], c.Status, c.Image})
	}
	display.Flush()
}
