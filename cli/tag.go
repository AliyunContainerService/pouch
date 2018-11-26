package main

import (
	"context"

	"github.com/spf13/cobra"
)

// tagDescription
var tagDescription = "tag command is to add tag reference for the existing image."

// TagCommand use to implement 'tag' command.
type TagCommand struct {
	baseCommand
}

// Init initialize tag command.
func (tag *TagCommand) Init(c *Cli) {
	tag.cli = c
	tag.cmd = &cobra.Command{
		Use:   "tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
		Short: "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE",
		Long:  tagDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return tag.runTag(args)
		},
		Example: tagExamples(),
	}
}

// runTag is the entry of tag command.
func (tag *TagCommand) runTag(args []string) error {
	ctx := context.Background()
	apiClient := tag.cli.Client()

	source, target := args[0], args[1]
	return apiClient.ImageTag(ctx, source, target)
}

// tagExamples shows examples in tag command, and is used in auto-generated cli docs.
func tagExamples() string {
	return `$ pouch tag registry.hub.docker.com/library/busybox:1.28 busybox:latest`
}
