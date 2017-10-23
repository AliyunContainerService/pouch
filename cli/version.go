package main

import (
	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// VersionCommand use to implement 'version' command.
type VersionCommand struct {
	baseCommand
}

// Init initialize version command.
func (v *VersionCommand) Init(c *Cli) {
	v.cli = c

	v.cmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
	}
}

// Run is the entry of version command.
func (v *VersionCommand) Run(args []string) {
	req, err := v.cli.NewGetRequest("/version")
	if err != nil {
		logrus.Errorln(err)
		return
	}
	resp := req.Send()

	if err := resp.Error(); err != nil {
		logrus.Errorf("failed to print version: %v, %s(%d)", err, resp.Status, resp.StatusCode)
		return
	}

	obj := &types.SystemVersion{}
	if err := resp.DecodeBody(obj); err != nil {
		logrus.Errorln(err)
		return
	}

	v.cli.Print(obj)
}
