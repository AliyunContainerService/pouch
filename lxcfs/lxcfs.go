package lxcfs

var (
	// IsLxcfsEnabled Whether to enable lxcfs
	IsLxcfsEnabled bool

	// LxcfsHomeDir is the absolute path of lxcfs
	LxcfsHomeDir string

	// LxcfsParentDir is the absolute path of the parent directory of lxcfs
	LxcfsParentDir string

	// LxcfsProcFiles is the crucial files in procfs
	LxcfsProcFiles = []string{"uptime", "swaps", "stat", "diskstats", "meminfo", "cpuinfo"}
)
