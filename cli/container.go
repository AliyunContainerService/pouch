package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

type container struct {
	name    string
	tty     bool
	volume  []string
	runtime string
	env     []string
	labels  []string
}

func (c *container) config() (*types.ContainerCreateConfig, error) {
	config := &types.ContainerCreateConfig{
		HostConfig: &types.HostConfig{},
	}

	// TODO
	config.Tty = c.tty
	config.Env = c.env

	// set labels
	config.Labels = make(map[string]string)
	for _, label := range c.labels {
		fields := strings.SplitN(label, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid label: %s", label)
		}
		k, v := fields[0], fields[1]
		config.Labels[k] = v
	}

	// set bind volume
	if c.volume != nil {
		config.HostConfig.Binds = c.volume
	}

	// set runtime
	if c.runtime != "" {
		config.HostConfig.Runtime = c.runtime
	}

	return config, nil
}
