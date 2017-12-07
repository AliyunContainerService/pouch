package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// imagesDescription is used to describe image command in detail and auto generate command doc.
// TODO: add description
var imagesDescription = ""

// ImageCommand use to implement 'images' command.
type ImageCommand struct {
	baseCommand

	// flags for image command
	flagQuiet  bool
	flagDigest bool
}

// Init initialize images command.
func (i *ImageCommand) Init(c *Cli) {
	i.cli = c
	i.cmd = &cobra.Command{
		Use:   "images",
		Short: "List all images",
		Long:  imagesDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.runImages(args)
		},
		Example: imagesExample(),
	}

	i.addFlags()
}

// addFlags adds flags for specific command.
func (i *ImageCommand) addFlags() {
	flagSet := i.cmd.Flags()
	flagSet.BoolVarP(&i.flagQuiet, "quiet", "q", false, "Only show image numeric ID")
	flagSet.BoolVar(&i.flagDigest, "digest", false, "Show images with digest")
}

// runImages is the entry of images container command.
func (i *ImageCommand) runImages(args []string) error {
	apiClient := i.cli.Client()

	imageList, err := apiClient.ImageList()
	if err != nil {
		return fmt.Errorf("failed to get image list: %v", err)

	}

	if i.flagQuiet {
		for _, image := range imageList {
			fmt.Println(image.ID)
		}
		return nil
	}

	if i.flagDigest {
		fmt.Printf("%-20s %-56s %-71s %s\n", "IMAGE ID", "IMAGE NAME", "DIGEST", "SIZE")
	} else {
		fmt.Printf("%-20s %-56s %s\n", "IMAGE ID", "IMAGE NAME", "SIZE")
	}

	for _, image := range imageList {
		if i.flagDigest {
			fmt.Printf("%-20s %-56s %-71s %d\n", image.ID, image.Name, image.Digest, image.Size)
		} else {
			fmt.Printf("%-20s %-56s %d\n", image.ID, image.Name, image.Size)
		}
	}
	return nil
}

// imagesExample shows examples in images command, and is used in auto-generated cli docs.
// TODO: add example
func imagesExample() string {
	return ""
}
