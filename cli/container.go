package main

import (
	"strings"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/opts/config"
	"github.com/alibaba/pouch/apis/types"

	strfmt "github.com/go-openapi/strfmt"
)

type container struct {
	labels              []string
	name                string
	tty                 bool
	volume              []string
	volumesFrom         []string
	runtime             string
	env                 []string
	entrypoint          string
	workdir             string
	user                string
	groupAdd            []string
	hostname            string
	rm                  bool
	disableNetworkFiles bool

	blkioWeight          uint16
	blkioWeightDevice    config.WeightDevice
	blkioDeviceReadBps   config.ThrottleBpsDevice
	blkioDeviceWriteBps  config.ThrottleBpsDevice
	blkioDeviceReadIOps  config.ThrottleIOpsDevice
	blkioDeviceWriteIOps config.ThrottleIOpsDevice

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
	ulimit         config.Ulimit
	pidsLimit      int64
	shmSize        string
	netPriority    int64

	// log driver and log option
	logDriver string
	logOpts   []string

	//add for rich container mode
	rich       bool
	richMode   string
	initScript string

	// nvidia container
	nvidiaVisibleDevices     string
	nvidiaDriverCapabilities string
}

func (c *container) config() (*types.ContainerCreateConfig, error) {
	labels, err := opts.ParseLabels(c.labels)
	if err != nil {
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

	logOpts, err := opts.ParseLogOptions(c.logDriver, c.logOpts)
	if err != nil {
		return nil, err
	}
	shmSize, err := opts.ParseShmSize(c.shmSize)
	if err != nil {
		return nil, err
	}

	config := &types.ContainerCreateConfig{
		ContainerConfig: types.ContainerConfig{
			Tty:                 c.tty,
			Env:                 c.env,
			Entrypoint:          strings.Fields(c.entrypoint),
			WorkingDir:          c.workdir,
			User:                c.user,
			Hostname:            strfmt.Hostname(c.hostname),
			DisableNetworkFiles: c.disableNetworkFiles,
			Labels:              labels,
			Rich:                c.rich,
			RichMode:            c.richMode,
			InitScript:          c.initScript,
			ExposedPorts:        ports,
			DiskQuota:           diskQuota,
			QuotaID:             c.quotaID,
			SpecAnnotation:      specAnnotation,
			NetPriority:         c.netPriority,
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
				BlkioWeightDevice:    c.blkioWeightDevice.Value(),
				BlkioDeviceReadBps:   c.blkioDeviceReadBps.Value(),
				BlkioDeviceReadIOps:  c.blkioDeviceReadIOps.Value(),
				BlkioDeviceWriteBps:  c.blkioDeviceWriteBps.Value(),
				BlkioDeviceWriteIOps: c.blkioDeviceWriteIOps.Value(),

				Devices:       deviceMappings,
				IntelRdtL3Cbm: intelRdtL3Cbm,
				CgroupParent:  c.cgroupParent,
				Ulimits:       c.ulimit.Value(),
				PidsLimit:     c.pidsLimit,
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
			LogConfig: &types.LogConfig{
				LogDriver: c.logDriver,
				LogOpts:   logOpts,
			},
			ShmSize: &shmSize,
		},

		NetworkingConfig: networkingConfig,
	}

	if c.nvidiaDriverCapabilities != "" || c.nvidiaVisibleDevices != "" {
		config.HostConfig.Resources.NvidiaConfig = &types.NvidiaConfig{
			NvidiaDriverCapabilities: c.nvidiaDriverCapabilities,
			NvidiaVisibleDevices:     c.nvidiaVisibleDevices,
		}
	}

	return config, nil
}
