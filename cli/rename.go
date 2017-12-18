package main

import (
	"github.com/spf13/cobra"
)

// renameDescription is used to describe rename command in detail and auto generate command doc.
var renameDescription = "Rename a container object in Pouchd. " +
	"You can change the name of one container identified by its name or ID. " +
	"The container you renamed is ready to be used by its new name."

// RenameCommand uses to implement 'rename' command, it renames a container.
type RenameCommand struct {
	baseCommand
}

// Init initialize rename command.
func (rc *RenameCommand) Init(c *Cli) {
	rc.cli = c

	rc.cmd = &cobra.Command{
		Use:   "rename [container] [newName]",
		Short: "Rename a container with newName",
		Long:  renameDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runRename(args)
		},
		Example: renameExample(),
	}
}

// runRename is the entry of rename command.
func (rc *RenameCommand) runRename(args []string) error {
	apiClient := rc.cli.Client()
	container := args[0]
	newName := args[1]

	err := apiClient.ContainerRename(container, newName)

	return err
}

// renameExample shows examples in rename command, and is used in auto-generated cli docs.
func renameExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
$ pouch rename foo newName
$ pouch ps
Name     ID       Status    Image                              Runtime
newName  71b9c1   Running   docker.io/library/busybox:latest   runc
`
}
