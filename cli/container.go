package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/runconfig"

	units "github.com/docker/go-units"
	strfmt "github.com/go-openapi/strfmt"
)

type container struct {
	labels           []string
	name             string
	tty              bool
	volume           []string
	runtime          string
	env              []string
	entrypoint       string
	workdir          string
	hostname         string
	cpushare         int64
	cpusetcpus       string
	cpusetmems       string
	memory           string
	memorySwap       string
	memorySwappiness int64
	devices          []string
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

	config.Entrypoint = strings.Fields(c.entrypoint)
	config.WorkingDir = c.workdir
	config.Hostname = strfmt.Hostname(c.hostname)

	// cgroup
	config.HostConfig.CPUShares = c.cpushare
	config.HostConfig.CpusetCpus = c.cpusetcpus
	config.HostConfig.CpusetMems = c.cpusetmems

	if c.memorySwappiness != -1 && (c.memorySwappiness < 0 || c.memorySwappiness > 100) {
		return nil, fmt.Errorf("invalid memory swappiness: %d (it's range is 0-100)", c.memorySwappiness)
	}
	config.HostConfig.MemorySwappiness = &c.memorySwappiness

	if c.memory != "" {
		v, err := units.RAMInBytes(c.memory)
		if err != nil {
			return nil, err
		}
		config.HostConfig.Memory = v
	}
	if c.memorySwap != "" {
		if c.memorySwap == "-1" {
			config.HostConfig.MemorySwap = -1
		} else {
			v, err := units.RAMInBytes(c.memorySwap)
			if err != nil {
				return nil, err
			}
			config.HostConfig.MemorySwap = v
		}
	}
	// parse device mappings
	deviceMappings := []*types.DeviceMapping{}
	for _, device := range c.devices {
		deviceMapping, err := runconfig.ParseDevice(device)
		if err != nil {
			return nil, fmt.Errorf("parse devices error: %s", err)
		}
		if !runconfig.ValidDeviceMode(deviceMapping.CgroupPermissions) {
			return nil, fmt.Errorf("%s invalid device mode: %s", device, deviceMapping.CgroupPermissions)
		}
		deviceMappings = append(deviceMappings, deviceMapping)
	}
	config.HostConfig.Devices = deviceMappings

	return config, nil
}
