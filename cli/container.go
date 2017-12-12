package main

import (
	"github.com/alibaba/pouch/apis/types"
)

type container struct {
	name    string
	tty     bool
	volume  []string
	runtime string
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
