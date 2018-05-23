package main

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// loadDescription is used to describe load command in detail and auto generate command doc.
var loadDescription = "load a set of images by tar stream"

// LoadCommand use to implement 'load' command.
type LoadCommand struct {
	baseCommand
	input string
}

// Init initialize load command.
func (l *LoadCommand) Init(c *Cli) {
	l.cli = c
	l.cmd = &cobra.Command{
		Use:   "load [OPTIONS] [IMAGE_NAME]",
		Short: "load a set of images from a tar archive or STDIN",
		Long:  loadDescription,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return l.runLoad(args)
		},
		Example: loadExample(),
	}
	l.addFlags()
}

// addFlags adds flags for specific command.
func (l *LoadCommand) addFlags() {
	flagSet := l.cmd.Flags()
	flagSet.StringVarP(&l.input, "input", "i", "", "Read from tar archive file, instead of STDIN")
}

// runLoad is the entry of load command.
func (l *LoadCommand) runLoad(args []string) error {
	ctx := context.Background()
	apiClient := l.cli.Client()

	var (
		in        io.Reader = os.Stdin
		imageName           = ""
	)

	if l.input != "" {
		file, err := os.Open(l.input)
		if err != nil {
			return err
		}

		defer file.Close()
		in = file
	}

	if len(args) > 0 {
		imageName = args[0]
	}
	return apiClient.ImageLoad(ctx, imageName, in)
}

// loadExample shows examples in load command, and is used in auto-generated cli docs.
func loadExample() string {
	return `$ pouch load -i busybox.tar busybox`
}
