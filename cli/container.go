package main

import (
	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

type container struct {
	name string
}

func (c *container) init() *cobra.Command {
	cmd := &cobra.Command{}

	// TODO add flag
	cmd.Flags().StringVar(&c.name, "name", "", "specified the container's name")

	return cmd
}

func (c *container) config() *types.ContainerConfigWrapper {
	config := &types.ContainerConfigWrapper{
		ContainerConfig: &types.ContainerConfig{},
		HostConfig:      &types.HostConfig{},
	}

	// TODO

	return config
}
