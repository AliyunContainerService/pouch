package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/pkg/utils"

	digest "github.com/opencontainers/go-digest"
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

type displayImage struct {
	id     string
	name   string
	size   imageSize
	digest string
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
		Use:   "images [OPTIONS]",
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
	ctx := context.Background()
	apiClient := i.cli.Client()

	imageList, err := apiClient.ImageList(ctx)
	if err != nil {
		return fmt.Errorf("failed to get image list: %v", err)

	}

	if i.flagQuiet {
		for _, image := range imageList {
			fmt.Println(utils.TruncateID(image.ID))
		}
		return nil
	}

	display := i.cli.NewTableDisplay()
	if i.flagDigest {
		display.AddRow([]string{"IMAGE ID", "IMAGE NAME", "DIGEST", "SIZE"})
	} else {
		display.AddRow([]string{"IMAGE ID", "IMAGE NAME", "SIZE"})
	}

	dimgs := make([]displayImage, 0, len(imageList))
	for _, img := range imageList {
		dimgs = append(dimgs, imageInfoToDisplayImages(img)...)
	}

	for _, dimg := range dimgs {
		if i.flagDigest {
			display.AddRow([]string{dimg.id, dimg.name, dimg.digest, fmt.Sprintf("%s", dimg.size)})
		} else {
			display.AddRow([]string{dimg.id, dimg.name, fmt.Sprintf("%s", dimg.size)})
		}
	}

	display.Flush()
	return nil
}

func imageInfoToDisplayImages(img types.ImageInfo) []displayImage {
	dimgs := make([]displayImage, 0)

	nameTags := make(map[string][]string)
	digestIndexByName := make(map[string]digest.Digest)

	for _, repoTag := range img.RepoTags {
		namedRef, err := reference.Parse(repoTag)
		// ideally, it should be nil
		if err != nil {
			continue
		}

		if reference.IsNameTagged(namedRef) {
			taggedRef := namedRef.(reference.Tagged)
			nameTags[taggedRef.Name()] = append(nameTags[taggedRef.Name()], taggedRef.Tag())
		}
	}

	for _, repoDigest := range img.RepoDigests {
		namedRef, err := reference.Parse(repoDigest)
		// ideally, it should be nil
		if err != nil {
			continue
		}

		namedRef = reference.TrimTagForDigest(namedRef)
		if cdRef, ok := namedRef.(reference.CanonicalDigested); ok {
			digestIndexByName[cdRef.Name()] = cdRef.Digest()
		}
	}

	for name, tags := range nameTags {
		for _, tag := range tags {
			dimg := displayImage{
				id:   utils.TruncateID(img.ID),
				name: name + ":" + tag,
				size: imageSize(img.Size),
			}

			if dig, ok := digestIndexByName[name]; ok {
				dimg.digest = dig.String()
			} else {
				dimg.digest = "<none>"
			}
			dimgs = append(dimgs, dimg)
		}
	}

	// if there is no repo tags
	if len(dimgs) == 0 {
		for name, dig := range digestIndexByName {
			dimgs = append(dimgs, displayImage{
				id:     utils.TruncateID(img.ID),
				name:   name + "@" + dig.String(),
				digest: dig.String(),
				size:   imageSize(img.Size),
			})
		}

		// if there is no repo digests
		if len(dimgs) == 0 {
			dimgs = append(dimgs, displayImage{
				id:     utils.TruncateID(img.ID),
				name:   "<none>",
				digest: "<none>",
				size:   imageSize(img.Size),
			})
		}
	}
	return dimgs
}

// imagesExample shows examples in images command, and is used in auto-generated cli docs.
func imagesExample() string {
	return `$ pouch images
IMAGE ID             IMAGE NAME                                               SIZE
bbc3a0323522         docker.io/library/busybox:latest                         703.14 KB
b81f317384d7         docker.io/library/nginx:latest                           42.39 MB`
}
