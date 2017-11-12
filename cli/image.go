package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
		Short: "show images",
	}

	i.addFlags()
}

func (i *ImageCommand) addFlags() {
	i.cmd.Flags().BoolVarP(&i.flagQuiet, "quiet", "q", false, "only show image numeric id")
	i.cmd.Flags().BoolVar(&i.flagDigest, "digest", false, "show image with digest")
}

// Run is the entry of images container command.
func (i *ImageCommand) Run(args []string) {
	apiClient := i.cli.Client()

	imageList, err := apiClient.ImageList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get image list: %s\n", err)
		return
	}

	if i.flagQuiet {
		for _, image := range imageList {
			fmt.Println(image.Name)
		}
		return
	}

	if i.flagDigest {
		fmt.Printf("%-20s %-56s %-71s %s\n", "IMAGE ID", "IMAGE NAME", "DIGEST", "SIZE")
	} else {
		fmt.Printf("%-20s %-56s %s\n", "IMAGE ID", "IMAGE NAME", "SIZE")
	}

	for _, image := range imageList {
		if i.flagDigest {
			fmt.Printf("%-20s %-56s %-71s %s\n", image.ID, image.Name, image.Digest, image.Size)
		} else {
			fmt.Printf("%-20s %-56s %s\n", image.ID, image.Name, image.Size)
		}
	}
}
