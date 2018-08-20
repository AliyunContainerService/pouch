package main

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// saveDescription is used to describe save command in detail and auto generate command doc.
var saveDescription = "save an image to a tar archive."

// SaveCommand use to implement 'save' command.
type SaveCommand struct {
	baseCommand
	output string
}

// Init initialize save command.
func (save *SaveCommand) Init(c *Cli) {
	save.cli = c
	save.cmd = &cobra.Command{
		Use:   "save [OPTIONS] IMAGE",
		Short: "Save an image to a tar archive or STDOUT",
		Long:  saveDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return save.runSave(args)
		},
		Example: saveExample(),
	}
	save.addFlags()
}

// addFlags adds flags for specific command.
func (save *SaveCommand) addFlags() {
	flagSet := save.cmd.Flags()
	flagSet.StringVarP(&save.output, "output", "o", "", "Save to a tar archive file, instead of STDOUT")
}

// runSave is the entry of save command.
func (save *SaveCommand) runSave(args []string) error {
	ctx := context.Background()
	apiClient := save.cli.Client()

	r, err := apiClient.ImageSave(ctx, args[0])
	if err != nil {
		return err
	}
	defer r.Close()

	out := os.Stdout
	if save.output != "" {
		out, err = os.Create(save.output)
		if err != nil {
			return nil
		}
		defer out.Close()
	}

	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

// saveExample shows examples in save command, and is used in auto-generated cli docs.
func saveExample() string {
	return `$ pouch save -o busybox.tar busybox:latest
$ pouch load -i busybox.tar foo
$ pouch images
IMAGE ID       IMAGE NAME                                           SIZE
8c811b4aec35   registry.hub.docker.com/library/busybox:latest       710.81 KB
8c811b4aec35   foo:latest                                           710.81 KB
`
}
