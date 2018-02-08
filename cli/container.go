package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/runconfig"

	units "github.com/docker/go-units"
	strfmt "github.com/go-openapi/strfmt"
)

type container struct {
	labels               []string
	name                 string
	tty                  bool
	volume               []string
	runtime              string
	env                  []string
	entrypoint           string
	workdir              string
	user                 string
	hostname             string
	cpushare             int64
	cpusetcpus           string
	cpusetmems           string
	memory               string
	memorySwap           string
	memorySwappiness     int64
	devices              []string
	enableLxcfs          bool
	privileged           bool
	restartPolicy        string
	ipcMode              string
	pidMode              string
	utsMode              string
	sysctls              []string
	networks             []string
	securityOpt          []string
	capAdd               []string
	capDrop              []string
	blkioWeight          uint16
	blkioWeightDevice    WeightDevice
	blkioDeviceReadBps   ThrottleBpsDevice
	blkioDeviceWriteBps  ThrottleBpsDevice
	blkioDeviceReadIOps  ThrottleIOpsDevice
	blkioDeviceWriteIOps ThrottleIOpsDevice
}

func (c *container) config() (*types.ContainerCreateConfig, error) {
	labels, err := parseLabels(c.labels)
	if err != nil {
		return nil, err
	}

	if err := validateMemorySwappiness(c.memorySwappiness); err != nil {
		return nil, err
	}

	memory, err := parseMemory(c.memory)
	if err != nil {
		return nil, err
	}

	memorySwap, err := parseMemorySwap(c.memorySwap)
	if err != nil {
		return nil, err
	}

	deviceMappings, err := parseDeviceMappings(c.devices)
	if err != nil {
		return nil, err
	}

	restartPolicy, err := parseRestartPolicy(c.restartPolicy)
	if err != nil {
		return nil, err
	}

	sysctls, err := parseSysctls(c.sysctls)
	if err != nil {
		return nil, err
	}

	var networkMode string
	// FIXME: Temporarily closed.
	//if len(c.networks) == 0 {
	//	networkMode = "bridge"
	//}
	networkingConfig := &types.NetworkingConfig{
		EndpointsConfig: map[string]*types.EndpointSettings{},
	}
	for _, network := range c.networks {
		name, parameter, mode, err := parseNetwork(network)
		if err != nil {
			return nil, err
		}

		if networkMode == "" || mode == "mode" {
			networkMode = name
		}

		if name == "container" {
			networkMode = fmt.Sprintf("%s:%s", name, parameter)
		} else if ipaddr := net.ParseIP(parameter); ipaddr != nil {
			networkingConfig.EndpointsConfig[name] = &types.EndpointSettings{
				IPAddress: parameter,
				IPAMConfig: &types.EndpointIPAMConfig{
					IPV4Address: parameter,
				},
			}
		}
	}

	config := &types.ContainerCreateConfig{
		ContainerConfig: types.ContainerConfig{
			Tty:        c.tty,
			Env:        c.env,
			Entrypoint: strings.Fields(c.entrypoint),
			WorkingDir: c.workdir,
			User:       c.user,
			Hostname:   strfmt.Hostname(c.hostname),
			Labels:     labels,
		},

		HostConfig: &types.HostConfig{
			Binds:   c.volume,
			Runtime: c.runtime,
			Resources: types.Resources{
				CPUShares:            c.cpushare,
				CpusetCpus:           c.cpusetcpus,
				CpusetMems:           c.cpusetmems,
				Devices:              deviceMappings,
				Memory:               memory,
				MemorySwap:           memorySwap,
				MemorySwappiness:     &c.memorySwappiness,
				BlkioWeight:          c.blkioWeight,
				BlkioWeightDevice:    c.blkioWeightDevice.value(),
				BlkioDeviceReadBps:   c.blkioDeviceReadBps.value(),
				BlkioDeviceReadIOps:  c.blkioDeviceReadIOps.value(),
				BlkioDeviceWriteBps:  c.blkioDeviceWriteBps.value(),
				BlkioDeviceWriteIOps: c.blkioDeviceWriteIOps.value(),
			},
			EnableLxcfs:   c.enableLxcfs,
			Privileged:    c.privileged,
			RestartPolicy: restartPolicy,
			IpcMode:       c.ipcMode,
			PidMode:       c.pidMode,
			UTSMode:       c.utsMode,
			Sysctls:       sysctls,
			SecurityOpt:   c.securityOpt,
			NetworkMode:   networkMode,
			CapAdd:        c.capAdd,
			CapDrop:       c.capDrop,
		},

		NetworkingConfig: networkingConfig,
	}

	return config, nil
}

func parseSysctls(sysctls []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, sysctl := range sysctls {
		fields := strings.SplitN(sysctl, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid sysctl: %s: sysctl must be in format of key=value", sysctl)
		}
		k, v := fields[0], fields[1]
		results[k] = v
	}
	return results, nil
}

func parseLabels(labels []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, label := range labels {
		fields := strings.SplitN(label, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid label: %s", label)
		}
		k, v := fields[0], fields[1]
		results[k] = v
	}
	return results, nil
}

func parseDeviceMappings(devices []string) ([]*types.DeviceMapping, error) {
	results := []*types.DeviceMapping{}
	for _, device := range devices {
		deviceMapping, err := runconfig.ParseDevice(device)
		if err != nil {
			return nil, fmt.Errorf("parse devices error: %s", err)
		}
		if !runconfig.ValidDeviceMode(deviceMapping.CgroupPermissions) {
			return nil, fmt.Errorf("%s invalid device mode: %s", device, deviceMapping.CgroupPermissions)
		}
		results = append(results, deviceMapping)
	}
	return results, nil
}

func parseMemory(memory string) (int64, error) {
	if memory == "" {
		return 0, nil
	}
	result, err := units.RAMInBytes(memory)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func parseMemorySwap(memorySwap string) (int64, error) {
	if memorySwap == "" {
		return 0, nil
	}
	if memorySwap == "-1" {
		return -1, nil
	}
	result, err := units.RAMInBytes(memorySwap)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func validateMemorySwappiness(memorySwappiness int64) error {
	if memorySwappiness != -1 && (memorySwappiness < 0 || memorySwappiness > 100) {
		return fmt.Errorf("invalid memory swappiness: %d (its range is -1 or 0-100)", memorySwappiness)
	}
	return nil
}

func parseRestartPolicy(restartPolicy string) (*types.RestartPolicy, error) {
	policy := &types.RestartPolicy{}

	if restartPolicy == "" {
		policy.Name = "no"
		return policy, nil
	}

	fields := strings.Split(restartPolicy, ":")
	policy.Name = fields[0]

	switch policy.Name {
	case "always", "unless-stopped", "no":
	case "on-failure":
		if len(fields) > 2 {
			return nil, fmt.Errorf("invalid restart policy: %s", restartPolicy)
		}
		if len(fields) == 2 {
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid restart policy: %v", err)
			}
			policy.MaximumRetryCount = int64(n)
		}
	default:
		return nil, fmt.Errorf("invalid restart policy: %s", restartPolicy)
	}

	return policy, nil
}

// network format as below:
// [network]:[ip_address], such as: mynetwork:172.17.0.2 or mynetwork(ip alloc by ipam) or 172.17.0.2(default network is bridge)
// [network_mode]:[parameter], such as: host(use host network) or container:containerID(use exist container network)
// [network_mode]:[parameter]:mode, such as: mynetwork:172.17.0.2:mode(if the container has multi-networks, the network is the default network mode)
func parseNetwork(network string) (string, string, string, error) {
	var (
		name      string
		parameter string
		mode      string
	)
	if network == "" {
		return "", "", "", fmt.Errorf("invalid network: cannot be empty")
	}
	arr := strings.Split(network, ":")
	switch len(arr) {
	case 1:
		if ipaddr := net.ParseIP(arr[0]); ipaddr != nil {
			parameter = arr[0]
		} else {
			name = arr[0]
		}
	case 2:
		name = arr[0]
		if name == "container" {
			parameter = arr[1]
		} else if ipaddr := net.ParseIP(arr[1]); ipaddr != nil {
			parameter = arr[1]
		} else {
			mode = arr[1]
		}
	default:
		name = arr[0]
		parameter = arr[1]
		mode = arr[2]
	}

	return name, parameter, mode, nil
}
