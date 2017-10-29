package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

// CreateCommand use to implement 'create' command, it create a container.
type CreateCommand struct {
	container
	baseCommand
}

// Init initialize create command.
func (cc *CreateCommand) Init(c *Cli) {
	cc.cli = c

	cc.cmd = cc.init()
	cc.cmd.Use = "create [image]"
	cc.cmd.Short = "Create a new container with specified image"
	cc.cmd.Args = cobra.MinimumNArgs(1)
}

// Run is the entry of create command.
func (cc *CreateCommand) Run(args []string) {
	config := cc.config()
	config.Image = args[0]

	path := "/containers/create"
	if cc.container.name != "" {
		path = fmt.Sprintf("/containers/create?name=%s", cc.container.name)
	}

	req, err := cc.cli.NewPostRequest(path, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new post request: %v \n", err)
		return
	}

	respone := req.Send()
	if err := respone.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v \n", err)
		return
	}
	defer respone.Close()

	result := types.ContainerCreateResp{}
	if err := respone.DecodeBody(&result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode body: %v \n", err)
		return
	}

	if len(result.Warnings) != 0 {
		fmt.Printf("WARNING: %s \n", strings.Join(result.Warnings, "\n"))
	}
	fmt.Printf("container's id: %s, name: %s \n", result.ID, result.Name)
}
