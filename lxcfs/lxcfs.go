package lxcfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

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

// CheckLxcfsMount check if the the mount point of lxcfs exists
func CheckLxcfsMount() error {
	isMount := false
	f, err := os.Open("/proc/1/mountinfo")
	if err != nil {
		return fmt.Errorf("Check lxcfs mounts failed: %v", err)
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	for {
		line, err := fr.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Check lxcfs mounts failed: %v", err)
		}

		if bytes.Contains(line, []byte(LxcfsHomeDir)) {
			isMount = true
			break
		}
	}
	if isMount == false {
		return fmt.Errorf("%s is not a mount point, please run \" lxcfs %s \" before Pouchd", LxcfsHomeDir, LxcfsHomeDir)
	}
	return nil
}
