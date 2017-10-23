package main

import (
	"github.com/spf13/cobra"
)

// SubCommand define some interfaces that the sub command must implement them.
type SubCommand interface {
	Init(*Cli)
	Cmd() *cobra.Command
	Run([]string)
}

type baseCommand struct {
	cmd *cobra.Command
	cli *Cli
}

func (b *baseCommand) Cmd() *cobra.Command {
	return b.cmd
}
