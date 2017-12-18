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
		HostConfig: &types.HostConfig{},
	}

	// TODO
	config.Tty = &c.tty

	// set bind volume
	if c.volume != nil {
		config.HostConfig.Binds = c.volume
	}

	// set runtime
	if c.runtime != "" {
		config.HostConfig.Runtime = c.runtime
	}

	return config
}
