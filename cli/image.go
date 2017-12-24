package main

import (
	"github.com/spf13/cobra"
)

// imageMgmtDescription is used to describe image command in detail and auto generate command doc.
var imageMgmtDescription = "Manage Pouch image"

// ImageMgmtCommand use to implement 'image' command.
type ImageMgmtCommand struct {
	baseCommand
}

// Init initialize "image" command.
func (i *ImageMgmtCommand) Init(c *Cli) {
	i.cli = c

	i.cmd = &cobra.Command{
		Use:   "image",
		Short: "Manage image",
		Long:  imageMgmtDescription,
		Args:  cobra.NoArgs,
	}

	i.cli.AddCommand(i, &ImageInspectCommand{})
}
