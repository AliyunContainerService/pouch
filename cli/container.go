package main

import (
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/opts"

	strfmt "github.com/go-openapi/strfmt"
)

type container struct {
	labels      []string
	name        string
	tty         bool
	volume      []string
	volumesFrom []string
	runtime     string
	env         []string
	entrypoint  string
	workdir     string
	user        string
	groupAdd    []string
	hostname    string
	rm          bool

	blkioWeight          uint16
	blkioWeightDevice    WeightDevice
	blkioDeviceReadBps   ThrottleBpsDevice
	blkioDeviceWriteBps  ThrottleBpsDevice
	blkioDeviceReadIOps  ThrottleIOpsDevice
	blkioDeviceWriteIOps ThrottleIOpsDevice

	cpushare   int64
	cpusetcpus string
	cpusetmems string
	cpuperiod  int64
	cpuquota   int64

	memory           string
	memorySwap       string
	memorySwappiness int64

	memoryWmarkRatio    int64
	memoryExtra         int64
	memoryForceEmptyCtl int64
	scheLatSwitch       int64
	oomKillDisable      bool

	devices        []string
	enableLxcfs    bool
	privileged     bool
	restartPolicy  string
	ipcMode        string
	pidMode        string
	utsMode        string
	sysctls        []string
	networks       []string
	ports          []string
	expose         []string
	publicAll      bool
	securityOpt    []string
	capAdd         []string
	capDrop        []string
	IntelRdtL3Cbm  string
	diskQuota      []string
	quotaID        string
	oomScoreAdj    int64
	specAnnotation []string
	cgroupParent   string

	//add for rich container mode
	rich       bool
	richMode   string
	initScript string
}

func (c *container) config() (*types.ContainerCreateConfig, error) {
	labels, err := opts.ParseLabels(c.labels)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateMemorySwappiness(c.memorySwappiness); err != nil {
		return nil, err
	}

	memory, err := opts.ParseMemory(c.memory)
	if err != nil {
		return nil, err
	}

	memorySwap, err := opts.ParseMemorySwap(c.memorySwap)
	if err != nil {
		return nil, err
	}

	intelRdtL3Cbm, err := opts.ParseIntelRdt(c.IntelRdtL3Cbm)
	if err != nil {
		return nil, err
	}

	deviceMappings, err := opts.ParseDeviceMappings(c.devices)
	if err != nil {
		return nil, err
	}

	restartPolicy, err := opts.ParseRestartPolicy(c.restartPolicy)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateRestartPolicy(restartPolicy); err != nil {
		return nil, err
	}

	sysctls, err := opts.ParseSysctls(c.sysctls)
	if err != nil {
		return nil, err
	}

	diskQuota, err := opts.ParseDiskQuota(c.diskQuota)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateDiskQuota(diskQuota); err != nil {
		return nil, err
	}

	specAnnotation, err := opts.ParseAnnotation(c.specAnnotation)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateOOMScore(c.oomScoreAdj); err != nil {
		return nil, err
	}

	if err := opts.ValidateCPUPeriod(c.cpuperiod); err != nil {
		return nil, err
	}

	if err := opts.ValidateCPUQuota(c.cpuquota); err != nil {
		return nil, err
	}

	networkingConfig, networkMode, err := opts.ParseNetworks(c.networks)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateNetworks(networkingConfig); err != nil {
		return nil, err
	}

	portBindings, err := opts.ParsePortBinding(c.ports)
	if err != nil {
		return nil, err
	}

	// FIXME(ziren): do we need verify portBinding ???
	if err := opts.ValidatePortBinding(portBindings); err != nil {
		return nil, err
	}

	ports, err := opts.ParseExposedPorts(c.ports, c.expose)
	if err != nil {
		return nil, err
	}

	config := &types.ContainerCreateConfig{
		ContainerConfig: types.ContainerConfig{
			Tty:            c.tty,
			Env:            c.env,
			Entrypoint:     strings.Fields(c.entrypoint),
			WorkingDir:     c.workdir,
			User:           c.user,
			Hostname:       strfmt.Hostname(c.hostname),
			Labels:         labels,
			Rich:           c.rich,
			RichMode:       c.richMode,
			InitScript:     c.initScript,
			ExposedPorts:   ports,
			DiskQuota:      diskQuota,
			QuotaID:        c.quotaID,
			SpecAnnotation: specAnnotation,
		},

		HostConfig: &types.HostConfig{
			Binds:       c.volume,
			VolumesFrom: c.volumesFrom,
			Runtime:     c.runtime,
			Resources: types.Resources{
				// cpu
				CPUShares:  c.cpushare,
				CpusetCpus: c.cpusetcpus,
				CpusetMems: c.cpusetmems,
				CPUPeriod:  c.cpuperiod,
				CPUQuota:   c.cpuquota,

				// memory
				Memory:           memory,
				MemorySwap:       memorySwap,
				MemorySwappiness: &c.memorySwappiness,
				// FIXME: validate in client side
				MemoryWmarkRatio:    &c.memoryWmarkRatio,
				MemoryExtra:         &c.memoryExtra,
				MemoryForceEmptyCtl: c.memoryForceEmptyCtl,
				ScheLatSwitch:       c.scheLatSwitch,
				OomKillDisable:      &c.oomKillDisable,

				// blkio
				BlkioWeight:          c.blkioWeight,
				BlkioWeightDevice:    c.blkioWeightDevice.value(),
				BlkioDeviceReadBps:   c.blkioDeviceReadBps.value(),
				BlkioDeviceReadIOps:  c.blkioDeviceReadIOps.value(),
				BlkioDeviceWriteBps:  c.blkioDeviceWriteBps.value(),
				BlkioDeviceWriteIOps: c.blkioDeviceWriteIOps.value(),

				Devices:       deviceMappings,
				IntelRdtL3Cbm: intelRdtL3Cbm,
				CgroupParent:  c.cgroupParent,
			},
			EnableLxcfs:   c.enableLxcfs,
			Privileged:    c.privileged,
			RestartPolicy: restartPolicy,
			IpcMode:       c.ipcMode,
			PidMode:       c.pidMode,
			UTSMode:       c.utsMode,
			GroupAdd:      c.groupAdd,
			Sysctls:       sysctls,
			SecurityOpt:   c.securityOpt,
			NetworkMode:   networkMode,
			CapAdd:        c.capAdd,
			CapDrop:       c.capDrop,
			PortBindings:  portBindings,
			OomScoreAdj:   c.oomScoreAdj,
		},

		NetworkingConfig: networkingConfig,
	}

	return config, nil
}
