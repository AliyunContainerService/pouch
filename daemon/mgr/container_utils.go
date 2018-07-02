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
func validateConfig(config *types.ContainerConfig, hostConfig *types.HostConfig, update bool) ([]string, error) {
	// validates container hostconfig
	warnings := make([]string, 0)
	warns, err := validateResource(&hostConfig.Resources, update)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, warns...)

	if hostConfig.OomScoreAdj < -1000 || hostConfig.OomScoreAdj > 1000 {
		return warnings, fmt.Errorf("oom score should be in range [-1000, 1000]")
	}

	if hostConfig.ShmSize != nil && *hostConfig.ShmSize < 0 {
		return warnings, fmt.Errorf("shm-size %d should greater than 0", *hostConfig.ShmSize)
	}

	// TODO: add more validate here
	return warnings, nil
}

func validateResource(r *types.Resources, update bool) ([]string, error) {
	cgroupInfo := system.NewCgroupInfo()
	if cgroupInfo == nil {
		return nil, nil
	}
	warnings := make([]string, 0, 64)

	// validates memory cgroup value
	if cgroupInfo.Memory != nil {
		if r.Memory > 0 && !cgroupInfo.Memory.MemoryLimit {
			logrus.Warn(MemoryWarn)
			warnings = append(warnings, MemoryWarn)
			r.Memory = 0
			r.MemorySwap = 0
		}
		if r.MemorySwap > 0 && !cgroupInfo.Memory.MemorySwap {
			logrus.Warn(MemorySwapWarn)
			warnings = append(warnings, MemorySwapWarn)
			r.MemorySwap = 0
		}
		// cgroup not allow memory-swap less than memory limit
		if r.Memory > 0 && r.MemorySwap > 0 && r.MemorySwap < r.Memory {
			return warnings, fmt.Errorf("Minimum memoryswap limit should be larger than memory limit")
		}
		// cgroup not allow set memory-swap without set memory
		if r.Memory == 0 && r.MemorySwap > 0 && !update {
			return warnings, fmt.Errorf("You should always set the Memory limit when using Memoryswap limit")
		}
		if r.Memory != 0 && r.Memory < MinMemory {
			return warnings, fmt.Errorf("Minimal memory should greater than 4M")
		}
		if r.Memory > 0 && r.MemorySwap > 0 && r.MemorySwap < 2*r.Memory {
			warnings = append(warnings, "You should typically size your swap space to approximately 2x main memory for systems with less than 2GB of RAM")
		}
		if r.MemorySwappiness != nil && !cgroupInfo.Memory.MemorySwappiness {
			logrus.Warn(MemorySwappinessWarn)
			warnings = append(warnings, MemorySwappinessWarn)
			r.MemorySwappiness = nil
		}
		if r.MemorySwappiness != nil && (*r.MemorySwappiness < 0 || *r.MemorySwappiness > 100) {
			return warnings, fmt.Errorf("MemorySwappiness should in range [0, 100]")
		}
		if r.OomKillDisable != nil && !cgroupInfo.Memory.OOMKillDisable {
			logrus.Warn(OOMKillWarn)
			warnings = append(warnings, OOMKillWarn)
			r.OomKillDisable = nil
		}
	}

	// validates cpu cgroup value
	if cgroupInfo.CPU != nil {
		if r.CpusetCpus != "" && !cgroupInfo.CPU.CpusetCpus {
			logrus.Warn(CpusetCpusWarn)
			warnings = append(warnings, CpusetCpusWarn)
			r.CpusetCpus = ""
		}
		if r.CpusetMems != "" && !cgroupInfo.CPU.CpusetMems {
			logrus.Warn(CpusetMemsWarn)
			warnings = append(warnings, CpusetMemsWarn)
			r.CpusetMems = ""
		}
		if r.CPUShares > 0 && !cgroupInfo.CPU.CPUShares {
			logrus.Warn(CPUSharesWarn)
			warnings = append(warnings, CPUSharesWarn)
			r.CPUShares = 0
		}
		if r.CPUQuota > 0 && !cgroupInfo.CPU.CPUQuota {
			logrus.Warn(CPUQuotaWarn)
			warnings = append(warnings, CPUQuotaWarn)
			r.CPUQuota = 0
		}
		// cpu.cfs_quota_us can accept value less than 0, we allow -1 and > 1000
		if r.CPUQuota > 0 && r.CPUQuota < 1000 {
			return warnings, fmt.Errorf("CPU cfs quota should be greater than 1ms(1000)")
		}
		if r.CPUPeriod > 0 && !cgroupInfo.CPU.CPUPeriod {
			logrus.Warn(CPUPeriodWarn)
			warnings = append(warnings, CPUPeriodWarn)
			r.CPUPeriod = 0
		}
		if r.CPUPeriod != 0 && (r.CPUPeriod < 1000 || r.CPUPeriod > 1000000) {
			return warnings, fmt.Errorf("CPU cfs period should be in range [1000, 1000000](1ms, 1s)")
		}
	}

	// validates blkio cgroup value
	if cgroupInfo.Blkio != nil {
		if r.BlkioWeight > 0 && !cgroupInfo.Blkio.BlkioWeight {
			logrus.Warn(BlkioWeightWarn)
			warnings = append(warnings, BlkioWeightWarn)
			r.BlkioWeight = 0
		}
		if len(r.BlkioWeightDevice) > 0 && !cgroupInfo.Blkio.BlkioWeightDevice {
			logrus.Warn(BlkioWeightDeviceWarn)
			warnings = append(warnings, BlkioWeightDeviceWarn)
			r.BlkioWeightDevice = []*types.WeightDevice{}
		}
		if len(r.BlkioDeviceReadBps) > 0 && !cgroupInfo.Blkio.BlkioDeviceReadBps {
			logrus.Warn(BlkioDeviceReadBpsWarn)
			warnings = append(warnings, BlkioDeviceReadBpsWarn)
			r.BlkioDeviceReadBps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceWriteBps) > 0 && !cgroupInfo.Blkio.BlkioDeviceWriteBps {
			logrus.Warn(BlkioDeviceWriteBpsWarn)
			warnings = append(warnings, BlkioDeviceWriteBpsWarn)
			r.BlkioDeviceWriteBps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceReadIOps) > 0 && !cgroupInfo.Blkio.BlkioDeviceReadIOps {
			logrus.Warn(BlkioDeviceReadIOpsWarn)
			warnings = append(warnings, BlkioDeviceReadIOpsWarn)
			r.BlkioDeviceReadIOps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceWriteIOps) > 0 && !cgroupInfo.Blkio.BlkioDeviceWriteIOps {
			logrus.Warn(BlkioDeviceWriteIOpsWarn)
			warnings = append(warnings, BlkioDeviceWriteIOpsWarn)
			r.BlkioDeviceWriteIOps = []*types.ThrottleDevice{}
		}
	}

	// validates pid cgroup value
	if cgroupInfo.Pids != nil {
		if r.PidsLimit != 0 && !cgroupInfo.Pids.Pids {
			logrus.Warn(PidsLimitWarn)
			warnings = append(warnings, PidsLimitWarn)
			r.PidsLimit = 0
		}
	}

	return warnings, nil
}

// amendContainerSettings modify config settings to wanted,
// it will be call before container created.
func amendContainerSettings(config *types.ContainerConfig, hostConfig *types.HostConfig) {
	r := &hostConfig.Resources
	if r.Memory > 0 && r.MemorySwap == 0 {
		r.MemorySwap = 2 * r.Memory
	}
}
