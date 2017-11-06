package main

import (
	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

type container struct {
	name   string
	tty    bool
	volume []string
}

func (c *container) init() *cobra.Command {
	cmd := &cobra.Command{}

	// TODO add flag
	cmd.Flags().StringVar(&c.name, "name", "", "specified the container's name")
	cmd.Flags().BoolVarP(&c.tty, "tty", "t", false, "allocate a tty device")
	cmd.Flags().StringSliceVarP(&c.volume, "volume", "v", nil, "create container with volumes")

	return cmd
}

func (c *container) config() *types.ContainerConfigWrapper {
	config := &types.ContainerConfigWrapper{
		ContainerConfig: &types.ContainerConfig{},
		HostConfig:      &types.HostConfig{},
	}

	// TODO
	config.Tty = &c.tty

	// set bind volume
	if c.volume != nil {
		config.HostConfig.Binds = c.volume
	}

	return config
}
