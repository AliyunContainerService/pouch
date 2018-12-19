package opts

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// CheckBind is used to check the volume bind information.
func CheckBind(b string) ([]string, error) {
	arr := strings.Split(b, ":")
	switch len(arr) {
	case 1:
		if arr[0] == "" {
			return nil, fmt.Errorf("unknown volume bind: %s", b)
		}
		if arr[0][:1] != "/" {
			return nil, fmt.Errorf("invalid bind path: %s", arr[0])
		}
	case 2, 3:
		if arr[1] == "" {
			return nil, fmt.Errorf("unknown volume bind: %s", b)
		}
		if arr[1][:1] != "/" {
			return nil, fmt.Errorf("invalid bind path: %s", arr[1])
		}
	default:
		return nil, fmt.Errorf("unknown volume bind: %s", b)
	}

	return arr, nil
}

// ParseVolumesFrom is used to parse the parameter of VolumesFrom.
func ParseVolumesFrom(volume string) (string, string, error) {
	if len(volume) == 0 {
		return "", "", fmt.Errorf("invalid argument volumes-from")
	}

	parts := strings.Split(volume, ":")
	containerID := parts[0]
	mode := ""
	if len(parts) > 1 {
		mode = parts[1]
	}

	if containerID == "" {
		return "", "", fmt.Errorf("failed to parse container's id")
	}

	return containerID, mode, nil
}

// ParseBindMode is used to parse the bind's mode.
func ParseBindMode(mp *types.MountPoint, mode string) error {
	mp.RW = true
	mp.CopyData = true

	defaultMode := 0
	rwMode := 0
	labelMode := 0
	replaceMode := 0
	copyMode := 0
	propagationMode := 0

	for _, m := range strings.Split(mode, ",") {
		switch m {
		case "":
			defaultMode++
		case "ro":
			mp.RW = false
			rwMode++
		case "rw":
			mp.RW = true
			rwMode++
		case "dr", "rr":
			// direct replace mode, random replace mode
			mp.Replace = m
			replaceMode++
		case "z", "Z":
			labelMode++
		case "nocopy":
			mp.CopyData = false
			copyMode++
		case "private", "rprivate", "slave", "rslave", "shared", "rshared":
			mp.Propagation = m
			propagationMode++
		default:
			return fmt.Errorf("unknown bind mode: %s", mode)
		}
	}

	if defaultMode > 1 || rwMode > 1 || replaceMode > 1 || copyMode > 1 || propagationMode > 1 {
		return fmt.Errorf("invalid bind mode: %s", mode)
	}

	if mode != "" {
		mp.Mode = mode
	}
	return nil
}

// CheckDuplicateMountPoint is used to check duplicate mount point
func CheckDuplicateMountPoint(mounts []*types.MountPoint, destination string) bool {
	destination = filepath.Clean(destination)
	for _, sm := range mounts {
		if filepath.Clean(sm.Destination) == destination {
			return true
		}
	}
	return false
}
