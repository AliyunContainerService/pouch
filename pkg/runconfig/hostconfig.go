package runconfig

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ParseDevice parses a device mapping string to a container.DeviceMapping struct
func ParseDevice(device string) (*types.DeviceMapping, error) {
	src := ""
	dst := ""
	permissions := "rwm"
	arr := strings.Split(device, ":")
	switch len(arr) {
	case 3:
		permissions = arr[2]
		fallthrough
	case 2:
		dst = arr[1]
		fallthrough
	case 1:
		src = arr[0]
	default:
		return nil, fmt.Errorf("invalid device specification: %s", device)
	}

	if dst == "" {
		dst = src
	}

	deviceMapping := &types.DeviceMapping{
		PathOnHost:        src,
		PathInContainer:   dst,
		CgroupPermissions: permissions,
	}
	return deviceMapping, nil
}

// ValidDeviceMode checks if the mode for device is valid or not.
// valid mode is a composition of r (read), w (write), and m (mknod).
func ValidDeviceMode(mode string) bool {
	var legalDeviceMode = map[rune]bool{
		'r': true,
		'w': true,
		'm': true,
	}
	if mode == "" {
		return false
	}
	for _, c := range mode {
		if !legalDeviceMode[c] {
			return false
		}
		legalDeviceMode[c] = false
	}
	return true
}
