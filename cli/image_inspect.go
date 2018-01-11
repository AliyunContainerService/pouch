package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alibaba/pouch/pkg/reference"
	"github.com/spf13/cobra"
)

// imageInspectDescription is used to describe inspect command in detail and auto generate command doc.
var imageInspectDescription = "Return detailed information on Pouch image"

// ImageInspectCommand use to implement 'image inspect' command.
type ImageInspectCommand struct {
	baseCommand
}

// Init initialize "image inspect" command.
func (i *ImageInspectCommand) Init(c *Cli) {
	i.cli = c
	i.cmd = &cobra.Command{
		Use:   "inspect [OPTIONS] IMAGE",
		Short: "Display detailed information on one image",
		Long:  imageInspectDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.runInspect(args)
		},
		Example: i.example(),
	}
}

// runInpsect is used to inspect image.
func (i *ImageInspectCommand) runInspect(args []string) error {
	ref, err := reference.Parse(args[0])
	if err != nil {
		return fmt.Errorf("failed to inspect image: %v", err)
	}

	apiClient := i.cli.Client()
	image, err := apiClient.ImageInspect(ref.String())
	if err != nil {
		return fmt.Errorf("failed to inspect image: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(image)
}

// example shows examples in inspect command, and is used in auto-generated cli docs.
func (i *ImageInspectCommand) example() string {
	return `$ pouch image inspect docker.io/library/busybox
{
  "CreatedAt": "2017-12-21 04:30:57",
  "Digest": "sha256:bbc3a03235220b170ba48a157dd097dd1379299370e1ed99ce976df0355d24f0",
  "ID": "bbc3a0323522",
  "Name": "docker.io/library/busybox:latest",
  "Size": 720019,
  "Tag": "latest"
}`
}
