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
	flagQuiet   bool
	flagDigest  bool
	flagNoTrunc bool
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
	flagSet.BoolVar(&i.flagNoTrunc, "no-trunc", false, "Do not truncate output")
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
			if i.flagNoTrunc {
				fmt.Println(image.ID)
			} else {
				fmt.Println(utils.TruncateID(image.ID))
			}
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
		dimgs = append(dimgs, imageInfoToDisplayImages(img, i.flagNoTrunc)...)
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

func imageInfoToDisplayImages(img types.ImageInfo, noTrunc bool) []displayImage {
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

	imageDisplayID := utils.TruncateID(img.ID)
	if noTrunc {
		imageDisplayID = img.ID
	}

	for name, tags := range nameTags {
		for _, tag := range tags {
			dimg := displayImage{
				id:   imageDisplayID,
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
				id:     imageDisplayID,
				name:   name + "@" + dig.String(),
				digest: dig.String(),
				size:   imageSize(img.Size),
			})
		}

		// if there is no repo digests
		if len(dimgs) == 0 {
			dimgs = append(dimgs, displayImage{
				id:     imageDisplayID,
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
b81f317384d7         docker.io/library/nginx:latest                           42.39 MB

$ pouch images --digest
IMAGE ID       IMAGE NAME                                           DIGEST                                                                    SIZE
2cb0d9787c4d   registry.hub.docker.com/library/hello-world:latest   sha256:4b8ff392a12ed9ea17784bd3c9a8b1fa3299cac44aca35a85c90c5e3c7afacdc   6.30 KB
4ab4c602aa5e   registry.hub.docker.com/library/hello-world:linux    sha256:d5c7d767f5ba807f9b363aa4db87d75ab030404a670880e16aedff16f605484b   5.25 KB

$ pouch images --no-trunc
IMAGE ID                                                                  IMAGE NAME                                           SIZE
sha256:2cb0d9787c4dd17ef9eb03e512923bc4db10add190d3f84af63b744e353a9b34   registry.hub.docker.com/library/hello-world:latest   6.30 KB
sha256:4ab4c602aa5eed5528a6620ff18a1dc4faef0e1ab3a5eddeddb410714478c67f   registry.hub.docker.com/library/hello-world:linux    5.25 KB`
}
