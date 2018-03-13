package main

import (
	"github.com/spf13/cobra"
)

// remountLxcfsDescription is used to describe remount-lxcfs command in detail and auto generate command doc.
var remountLxcfsDescription = "\nremount lxcfs in containers."

// RemountLxcfsCommand is used to implement 'ps' command.
type RemountLxcfsCommand struct {
	baseCommand
}

// Init initializes remountLxcfsCommand command.
func (p *RemountLxcfsCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "remount-lxcfs",
		Short: "remount lxcfs bind in containers",
		Long:  remountLxcfsDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runRemountLxcfs(args)
		},
		Example: remountLxcfsExample(),
	}
}

// runRemountLxcfs is the entry of remountLxcfsCommand command.
func (p *RemountLxcfsCommand) runRemountLxcfs(args []string) error {
	// TODO: get containers which enable lxcfs, and exec remount in them
	return nil
}

// remountLxcfsExampleExample shows examples in remount-lxcfs command, and is used in auto-generated cli docs.
func remountLxcfsExample() string {
	return `$ pouch remount-lxcfs
ID       Status
e42c68   OK
`
}
