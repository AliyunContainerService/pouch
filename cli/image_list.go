package main

import (
	"fmt"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/spf13/cobra"
)

// imagesDescription is used to describe image command in detail and auto generate command doc.
var imagesDescription = "List all images in Pouchd." +
	"This is useful when you wish to have a look at images and Pouchd will show all local images with their NAME and SIZE." +
	"All local images will be shown in a table format you can use."

type imageSize int64

func (i imageSize) String() string {
	return utils.FormatSize(int64(i))
}

// ImagesCommand use to implement 'images' command.
type ImagesCommand struct {
	baseCommand

	// flags for image command
	flagQuiet  bool
	flagDigest bool
}

// Init initialize images command.
func (i *ImagesCommand) Init(c *Cli) {
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
func (i *ImagesCommand) addFlags() {
	flagSet := i.cmd.Flags()
	flagSet.BoolVarP(&i.flagQuiet, "quiet", "q", false, "Only show image numeric ID")
	flagSet.BoolVar(&i.flagDigest, "digest", false, "Show images with digest")
}

// runImages is the entry of images container command.
func (i *ImagesCommand) runImages(args []string) error {
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

	display := i.cli.NewTableDisplay()

	if i.flagDigest {
		display.AddRow([]string{"IMAGE ID", "IMAGE NAME", "DIGEST", "SIZE"})
	} else {
		display.AddRow([]string{"IMAGE ID", "IMAGE NAME", "SIZE"})
	}

	for _, image := range imageList {
		if i.flagDigest {
			display.AddRow([]string{
				image.ID,
				image.Name,
				image.Digest,
				fmt.Sprintf("%s", imageSize(image.Size)),
			})
		} else {
			display.AddRow([]string{
				image.ID,
				image.Name,
				fmt.Sprintf("%s", imageSize(image.Size)),
			})
		}
	}

	display.Flush()
	return nil
}

// imagesExample shows examples in images command, and is used in auto-generated cli docs.
func imagesExample() string {
	return `$ pouch images
IMAGE ID             IMAGE NAME                                               SIZE
bbc3a0323522         docker.io/library/busybox:latest                         703.14 KB
b81f317384d7         docker.io/library/nginx:latest                           42.39 MB`
}
