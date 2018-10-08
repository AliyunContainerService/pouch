package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/spf13/cobra"
)

// commitDescription is used to describe commit command in detail and auto generate command doc.
var commitDescription = "commit an image from a container."

// CommitCommand is used to implement 'commit' command.
type CommitCommand struct {
	baseCommand
	author  string
	message string
}

// Init initializes CommitCommand command.
func (cc *CommitCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "commit [OPTIONS] CONTAINER REPOSITORY[:TAG]",
		Short: "Commit an image from a container",
		Long:  commitDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCommit(args)
		},
		Example: commitExample(),
	}
	cc.addFlags()
}

// addFlags adds flags for specific command.
func (cc *CommitCommand) addFlags() {
	flagSet := cc.cmd.Flags()

	flagSet.StringVarP(&cc.author, "author", "a", "", "Image author, eg.(name <email@email.com>)")
	flagSet.StringVarP(&cc.message, "message", "m", "", "Commit message")
}

// runCommit is the entry of CommitCommand command.
func (cc *CommitCommand) runCommit(args []string) error {
	ctx := context.Background()
	apiClient := cc.cli.Client()

	// create commit process.
	id := args[0]
	ref := args[1]

	namedRef, err := reference.Parse(ref)
	if err != nil {
		return err
	}

	namedRef = reference.WithDefaultTagIfMissing(namedRef)

	var name, tag string
	if reference.IsNameTagged(namedRef) {
		name, tag = namedRef.Name(), namedRef.(reference.Tagged).Tag()
	} else {
		name, tag = namedRef.String(), "latest"
	}

	commitConfig := types.ContainerCommitOptions{
		Repository: name,
		Tag:        tag,
		Comment:    cc.message,
		Author:     cc.author,
	}

	respCommit, err := apiClient.ContainerCommit(ctx, id, commitConfig)
	if err != nil {
		return fmt.Errorf("failed to commit container %s: %v", id, err)
	}

	fmt.Fprintln(os.Stdout, respCommit.ID)
	return nil
}

// commitExample shows examples in commit command, and is used in auto-generated cli docs.
func commitExample() string {
	return `$ pouch commit 25bf50 test:image
1c7e415csa333
`
}
