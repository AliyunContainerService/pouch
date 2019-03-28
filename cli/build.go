package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/pouch/cli/build"

	"github.com/spf13/cobra"
)

// buildDescription is used to describe build command in detail and auto generate command doc.
var buildDescription = "Build an image from a Dockerfile"

// BuildCommand use to implement 'build' command, it download image.
type BuildCommand struct {
	baseCommand

	buildArgs []string
	tagList   []string
	target    string
	addr      string
}

// Init initialize pull command.
func (b *BuildCommand) Init(c *Cli) {
	b.cli = c

	b.cmd = &cobra.Command{
		Use:   "build [OPTION] PATH",
		Short: "Build an image from a Dockerfile",
		Long:  buildDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.runBuild(args)
		},
	}
	b.addFlags()
}

func (b *BuildCommand) addFlags() {
	flagSet := b.cmd.Flags()

	flagSet.StringArrayVar(&b.buildArgs, "build-arg", nil, "Set build-time variables")
	flagSet.StringArrayVarP(&b.tagList, "tag", "t", nil, "Name and optionally a tag in the 'name:tag' format")
	flagSet.StringVar(&b.target, "target", "", "Set the target build stage to build")
	flagSet.StringVar(&b.addr, "addr", "unix:///run/buildkit/buildkitd.sock", "buildkitd address")
}

func (b *BuildCommand) runBuild(args []string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	opts := b.buildOptions(args[0])
	return build.Build(ctx, b.addr, opts)
}

func (b *BuildCommand) buildOptions(workdir string) *build.Options {
	opts := &build.Options{
		TagList: b.tagList,
		// TODO: build args
		Target: b.target,
	}

	opts.LocalDirs = map[string]string{
		"dockerfile": workdir,
		"context":    workdir,
	}

	// using unknown:timestamp if there is no tag
	if len(opts.TagList) == 0 {
		opts.TagList = append(opts.TagList, fmt.Sprintf("unknown:%v", time.Now().UnixNano()))
	}
	return opts
}
