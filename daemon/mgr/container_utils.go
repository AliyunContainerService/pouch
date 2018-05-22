package mgr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/opencontainers/selinux/go-selinux/label"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// containerID returns the container's id, the parameter 'nameOrPrefix' may be container's
// name, id or prefix id.
func (mgr *ContainerManager) containerID(nameOrPrefix string) (string, error) {
	var obj meta.Object

	// name is the container's name.
	id, ok := mgr.NameToID.Get(nameOrPrefix).String()
	if ok {
		return id, nil
	}

	// name is the container's prefix of the id.
	objs, err := mgr.Store.GetWithPrefix(nameOrPrefix)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get container info with prefix: %s", nameOrPrefix)
	}
	if len(objs) > 1 {
		return "", errors.Wrap(errtypes.ErrTooMany, "container: "+nameOrPrefix)
	}
	if len(objs) == 0 {
		return "", errors.Wrap(errtypes.ErrNotfound, "container: "+nameOrPrefix)
	}
	obj = objs[0]

	con, ok := obj.(*Container)
	if !ok {
		return "", fmt.Errorf("failed to get container info, invalid meta's type")
	}

	return con.ID, nil
}

func (mgr *ContainerManager) container(nameOrPrefix string) (*Container, error) {
	res, ok := mgr.cache.Get(nameOrPrefix).Result()
	if ok {
		return res.(*Container), nil
	}

	id, err := mgr.containerID(nameOrPrefix)
	if err != nil {
		return nil, err
	}

	// lookup again
	res, ok = mgr.cache.Get(id).Result()
	if ok {
		return res.(*Container), nil
	}

	return nil, errors.Wrap(errtypes.ErrNotfound, "container "+nameOrPrefix)
}

// generateID generates an ID for newly created container. We must ensure that
// this ID has not used yet.
func (mgr *ContainerManager) generateID() (string, error) {
	var id string
	for {
		id = randomid.Generate()
		_, err := mgr.Store.Get(id)
		if err != nil {
			if merr, ok := err.(meta.Error); ok && merr.IsNotfound() {
				break
			}
			return "", err
		}
	}
	return id, nil
}

// generateName generates container name by container ID.
// It get first 6 continuous letters which has not been taken.
// TODO: take a shorter than 6 letters ID into consideration.
// FIXME: there is possibility that for block loops forever.
func (mgr *ContainerManager) generateName(id string) string {
	var name string
	i := 0
	for {
		if i+6 > len(id) {
			break
		}
		name = id[i : i+6]
		i++
		if !mgr.NameToID.Get(name).Exist() {
			break
		}
	}
	return name
}

// BuildContainerEndpoint is used to build container's endpoint config.
func BuildContainerEndpoint(c *Container) *networktypes.Endpoint {
	return &networktypes.Endpoint{
		Owner:           c.ID,
		Hostname:        c.Config.Hostname,
		Domainname:      c.Config.Domainname,
		HostsPath:       c.HostsPath,
		ExtraHosts:      c.HostConfig.ExtraHosts,
		HostnamePath:    c.HostnamePath,
		ResolvConfPath:  c.ResolvConfPath,
		NetworkDisabled: c.Config.NetworkDisabled,
		NetworkMode:     c.HostConfig.NetworkMode,
		DNS:             c.HostConfig.DNS,
		DNSOptions:      c.HostConfig.DNSOptions,
		DNSSearch:       c.HostConfig.DNSSearch,
		MacAddress:      c.Config.MacAddress,
		PublishAllPorts: c.HostConfig.PublishAllPorts,
		ExposedPorts:    c.Config.ExposedPorts,
		PortBindings:    c.HostConfig.PortBindings,
		NetworkConfig:   c.NetworkSettings,
	}
}

func parseSecurityOpts(c *Container, securityOpts []string) error {
	var (
		labelOpts []string
		err       error
	)
	for _, securityOpt := range securityOpts {
		if securityOpt == "no-new-privileges" {
			c.NoNewPrivileges = true
			continue
		}
		fields := strings.SplitN(securityOpt, "=", 2)
		if len(fields) != 2 {
			return fmt.Errorf("invalid --security-opt %s: must be in format of key=value", securityOpt)
		}
		key, value := fields[0], fields[1]
		switch key {
		// TODO: handle other security options.
		case "apparmor":
			c.AppArmorProfile = value
		case "seccomp":
			c.SeccompProfile = value
		case "no-new-privileges":
			noNewPrivileges, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid --security-opt: %q", securityOpt)
			}
			c.NoNewPrivileges = noNewPrivileges
		case "label":
			labelOpts = append(labelOpts, value)
		default:
			return fmt.Errorf("invalid type %s in --security-opt %s: unknown type from apparmor, seccomp, no-new-privileges and SELinux label", key, securityOpt)
		}
	}

	if len(labelOpts) == 0 {
		return nil
	}
	c.ProcessLabel, c.MountLabel, err = label.InitLabels(labelOpts)
	if err != nil {
		return fmt.Errorf("failed to init labels: %v", err)
	}

	return nil
}

// fieldsASCII is similar to strings.Fields but only allows ASCII whitespaces
func fieldsASCII(s string) []string {
	fn := func(r rune) bool {
		switch r {
		case '\t', '\n', '\f', '\r', ' ':
			return true
		}
		return false
	}
	return strings.FieldsFunc(s, fn)
}

func parsePSOutput(output []byte, pids []int) (*types.ContainerProcessList, error) {
	procList := &types.ContainerProcessList{}

	lines := strings.Split(string(output), "\n")
	procList.Titles = fieldsASCII(lines[0])

	pidIndex := -1
	for i, name := range procList.Titles {
		if name == "PID" {
			pidIndex = i
		}
	}
	if pidIndex == -1 {
		return nil, fmt.Errorf("Couldn't find PID field in ps output")
	}

	// loop through the output and extract the PID from each line
	for _, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		fields := fieldsASCII(line)
		p, err := strconv.Atoi(fields[pidIndex])
		if err != nil {
			return nil, fmt.Errorf("Unexpected pid '%s': %s", fields[pidIndex], err)
		}

		for _, pid := range pids {
			if pid == p {
				// Make sure number of fields equals number of header titles
				// merging "overhanging" fields
				process := fields[:len(procList.Titles)-1]
				process = append(process, strings.Join(fields[len(procList.Titles)-1:], " "))
				procList.Processes = append(procList.Processes, process)
			}
		}
	}
	return procList, nil
}

// validateConfig validates container config
func validateConfig(config *types.ContainerCreateConfig) ([]string, error) {
	// validates container hostconfig
	warnings := make([]string, 0)
	warns, err := validateResource(&config.HostConfig.Resources)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, warns...)

	// TODO: add more validate here
	return warnings, nil
}

func validateResource(r *types.Resources) ([]string, error) {
	cgroupInfo := system.NewCgroupInfo()
	if cgroupInfo == nil {
		return nil, nil
	}
	warnings := make([]string, 0, 64)

	// validates memory cgroup value
	if cgroupInfo.Memory != nil {
		if !cgroupInfo.Memory.MemoryLimit {
			warn := "Current Kernel does not support memory limit, discard --memory"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.Memory = 0
		}
		if !cgroupInfo.Memory.MemorySwap {
			warn := "Current Kernel does not support memory swap, discard --memory-swap"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.MemorySwap = -1
		}
		if !cgroupInfo.Memory.MemorySwappiness {
			warn := "Current Kernel does not support memory swappiness , discard --memory-swappiness"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.MemorySwappiness = nil
		}
		if !cgroupInfo.Memory.OOMKillDisable {
			warn := "Current Kernel does not support disable oom kill, discard --oom-kill-disable"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.OomKillDisable = nil
		}
	}

	// validates cpu cgroup value
	if cgroupInfo.CPU != nil {
		if !cgroupInfo.CPU.CpusetCpus {
			warn := "Current Kernel does not support cpuset cpus, discard --cpuset-cpus"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.CpusetCpus = ""
		}
		if !cgroupInfo.CPU.CpusetCpus {
			warn := "Current Kernel does not support cpuset cpus, discard --cpuset-mems"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.CpusetMems = ""
		}
		if !cgroupInfo.CPU.CPUShares {
			warn := "Current Kernel does not support cpu shares, discard --cpu-share"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.CPUShares = 0
		}
		if !cgroupInfo.CPU.CPUQuota {
			warn := "Current Kernel does not support cpu quota, discard --cpu-quota"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.CPUQuota = 0
		}
		if !cgroupInfo.CPU.CPUPeriod {
			warn := "Current Kernel does not support cpu period, discard --cpu-period"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.CPUPeriod = 0
		}
	}

	// validates blkio cgroup value
	if cgroupInfo.Blkio != nil {
		if !cgroupInfo.Blkio.BlkioWeight {
			warn := "Current Kernel does not support blkio weight, discard --blkio-weight"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioWeight = 0
		}
		if !cgroupInfo.Blkio.BlkioWeightDevice {
			warn := "Current Kernel does not support blkio weight device, discard --blkio-weight-device"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioWeightDevice = []*types.WeightDevice{}
		}
		if !cgroupInfo.Blkio.BlkioDeviceReadBps {
			warn := "Current Kernel does not support blkio device throttle read bps, discard --device-read-bps"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioDeviceReadBps = []*types.ThrottleDevice{}
		}
		if !cgroupInfo.Blkio.BlkioDeviceWriteBps {
			warn := "Current Kernel does not support blkio device throttle write bps, discard --device-write-bps"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioDeviceWriteBps = []*types.ThrottleDevice{}
		}
		if !cgroupInfo.Blkio.BlkioDeviceReadIOps {
			warn := "Current Kernel does not support blkio device throttle read iops, discard --device-read-iops"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioDeviceReadIOps = []*types.ThrottleDevice{}
		}
		if !cgroupInfo.Blkio.BlkioDeviceWriteIOps {
			warn := "Current Kernel does not support blkio device throttle, discard --device-write-iops"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.BlkioDeviceWriteIOps = []*types.ThrottleDevice{}
		}
	}

	// validates pid cgroup value
	if cgroupInfo.Pids != nil {
		if !cgroupInfo.Pids.Pids {
			warn := "Current Kernel does not support pids cgroup, discard --pids-limit"
			logrus.Warn(warn)
			warnings = append(warnings, warn)
			r.PidsLimit = 0
		}
	}

	return warnings, nil
}
