package main

import (
	"github.com/spf13/pflag"
)

func addCommonFlags(flagSet *pflag.FlagSet) *container {
	c := &container{}

	// please add the following flag by name in alphabetical order
	// blkio
	flagSet.Uint16Var(&c.blkioWeight, "blkio-weight", 0, "Block IO (relative weight), between 10 and 1000, or 0 to disable")
	flagSet.Var(&c.blkioWeightDevice, "blkio-weight-device", "Block IO weight (relative device weight), need CFQ IO Scheduler enable")
	flagSet.Var(&c.blkioDeviceReadBps, "device-read-bps", "Limit read rate (bytes per second) from a device")
	flagSet.Var(&c.blkioDeviceReadIOps, "device-read-iops", "Limit read rate (IO per second) from a device")
	flagSet.Var(&c.blkioDeviceWriteBps, "device-write-bps", "Limit write rate (bytes per second) from a device")
	flagSet.Var(&c.blkioDeviceWriteIOps, "device-write-iops", "Limit write rate (IO per second) from a device")

	// capbilities
	flagSet.StringSliceVar(&c.capAdd, "cap-add", nil, "Add Linux capabilities")
	flagSet.StringSliceVar(&c.capDrop, "cap-drop", nil, "Drop Linux capabilities")

	// cpu
	flagSet.Int64Var(&c.cpushare, "cpu-share", 0, "CPU shares (relative weight)")
	flagSet.StringVar(&c.cpusetcpus, "cpuset-cpus", "", "CPUs in which to allow execution (0-3, 0,1)")
	flagSet.StringVar(&c.cpusetmems, "cpuset-mems", "", "MEMs in which to allow execution (0-3, 0,1)")
	flagSet.Int64Var(&c.cpuperiod, "cpu-period", 0, "Limit CPU CFS (Completely Fair Scheduler) period, range is in [1000(1ms),1000000(1s)]")
	flagSet.Int64Var(&c.cpuquota, "cpu-quota", 0, "Limit CPU CFS (Completely Fair Scheduler) quota, range is in [1000,∞)")

	// device related options
	flagSet.StringSliceVarP(&c.devices, "device", "", nil, "Add a host device to the container")

	flagSet.BoolVar(&c.enableLxcfs, "enableLxcfs", false, "Enable lxcfs for the container, only effective when enable-lxcfs switched on in Pouchd")
	flagSet.StringVar(&c.entrypoint, "entrypoint", "", "Overwrite the default ENTRYPOINT of the image")
	flagSet.StringSliceVarP(&c.env, "env", "e", nil, "Set environment variables for container")
	flagSet.StringVar(&c.hostname, "hostname", "", "Set container's hostname")
	flagSet.BoolVar(&c.disableNetworkFiles, "disable-network-files", false, "Disable the generation of network files(/etc/hostname, /etc/hosts and /etc/resolv.conf) for container. If true, no network files will be generated. Default false")

	// Intel RDT
	flagSet.StringVar(&c.IntelRdtL3Cbm, "intel-rdt-l3-cbm", "", "Limit container resource for Intel RDT/CAT which introduced in Linux 4.10 kernel")

	flagSet.StringVar(&c.ipcMode, "ipc", "", "IPC namespace to use")
	flagSet.StringSliceVarP(&c.labels, "label", "l", nil, "Set labels for a container")

	// log driver and log options
	flagSet.StringVar(&c.logDriver, "log-driver", "json-file", "Logging driver for the container")
	flagSet.StringSliceVar(&c.logOpts, "log-opt", nil, "Log driver options")

	// memory

	flagSet.StringVarP(&c.memory, "memory", "m", "", "Memory limit")
	flagSet.StringVar(&c.memorySwap, "memory-swap", "", "Swap limit equal to memory + swap, '-1' to enable unlimited swap")
	flagSet.Int64Var(&c.memorySwappiness, "memory-swappiness", -1, "Container memory swappiness [0, 100]")
	// for alikernel isolation options
	flagSet.Int64Var(&c.memoryWmarkRatio, "memory-wmark-ratio", 0, "Represent this container's memory low water mark percentage, range in [0, 100]. The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio")
	flagSet.Int64Var(&c.memoryExtra, "memory-extra", 0, "Represent container's memory high water mark percentage, range in [0, 100]")
	flagSet.Int64Var(&c.memoryForceEmptyCtl, "memory-force-empty-ctl", 0, "Whether to reclaim page cache when deleting the cgroup of container")
	flagSet.BoolVar(&c.oomKillDisable, "oom-kill-disable", false, "Disable OOM Killer")
	flagSet.Int64Var(&c.oomScoreAdj, "oom-score-adj", -500, "Tune host's OOM preferences (-1000 to 1000)")

	flagSet.StringVar(&c.name, "name", "", "Specify name of container")

	flagSet.StringSliceVar(&c.networks, "net", nil, "Set networks to container")
	flagSet.StringSliceVarP(&c.ports, "port", "p", nil, "Set container ports mapping")
	flagSet.StringSliceVar(&c.expose, "expose", nil, "Set expose container's ports")

	flagSet.StringVar(&c.pidMode, "pid", "", "PID namespace to use")
	flagSet.BoolVar(&c.privileged, "privileged", false, "Give extended privileges to the container")

	flagSet.StringVar(&c.restartPolicy, "restart", "", "Restart policy to apply when container exits")
	flagSet.StringVar(&c.runtime, "runtime", "", "OCI runtime to use for this container")

	flagSet.StringSliceVar(&c.securityOpt, "security-opt", nil, "Security Options")

	flagSet.Int64Var(&c.scheLatSwitch, "sche-lat-switch", 0, "Whether to enable scheduler latency count in cpuacct")

	flagSet.StringSliceVar(&c.sysctls, "sysctl", nil, "Sysctl options")
	flagSet.BoolVarP(&c.tty, "tty", "t", false, "Allocate a pseudo-TTY")

	// user
	flagSet.StringVarP(&c.user, "user", "u", "", "UID")

	flagSet.StringSliceVar(&c.groupAdd, "group-add", nil, "Add additional groups to join")

	flagSet.StringVar(&c.utsMode, "uts", "", "UTS namespace to use")

	flagSet.StringSliceVarP(&c.volume, "volume", "v", nil, "Bind mount volumes to container, format is: [source:]<destination>[:mode], [source] can be volume or host's path, <destination> is container's path, [mode] can be \"ro/rw/dr/rr/z/Z/nocopy/private/rprivate/slave/rslave/shared/rshared\"")
	flagSet.StringSliceVar(&c.volumesFrom, "volumes-from", nil, "set volumes from other containers, format is <container>[:mode]")

	flagSet.StringVarP(&c.workdir, "workdir", "w", "", "Set the working directory in a container")
	flagSet.Var(&c.ulimit, "ulimit", "Set container ulimit")
	flagSet.Int64Var(&c.pidsLimit, "pids-limit", 0, "Set container pids limit")

	flagSet.BoolVar(&c.rich, "rich", false, "Start container in rich container mode. (default false)")
	flagSet.StringVar(&c.richMode, "rich-mode", "", "Choose one rich container mode. dumb-init(default), systemd, sbin-init")
	flagSet.StringVar(&c.initScript, "initscript", "", "Initial script executed in container")

	// cgroup
	flagSet.StringVarP(&c.cgroupParent, "cgroup-parent", "", "", "Optional parent cgroup for the container")

	// disk quota
	flagSet.StringSliceVar(&c.diskQuota, "disk-quota", nil, "Set disk quota for container")
	flagSet.StringVar(&c.quotaID, "quota-id", "", "Specified quota id, if id < 0, it means pouchd alloc a unique quota id")

	// additional runtime spec annotations
	flagSet.StringSliceVar(&c.specAnnotation, "annotation", nil, "Additional annotation for runtime")

	return c
}
