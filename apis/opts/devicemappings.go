package opts

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ParseDeviceMappings parse devicemappings
func ParseDeviceMappings(devices []string) ([]*types.DeviceMapping, error) {
	results := []*types.DeviceMapping{}
	for _, device := range devices {
		deviceMapping, err := parseDevice(device)
		if err != nil {
			return nil, fmt.Errorf("failed to parse devices: %v", err)
		}

		if !ValidateDeviceMode(deviceMapping.CgroupPermissions) {
			return nil, fmt.Errorf("%s invalid device mode: %s", device, deviceMapping.CgroupPermissions)
		}

		results = append(results, deviceMapping)
	}
	return results, nil

}

// parseDevice parses a device mapping string to a container.DeviceMapping struct
func parseDevice(device string) (*types.DeviceMapping, error) {
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

// ValidateDeviceMode checks if the mode for device is valid or not.
// valid mode is a composition of r (read), w (write), and m (mknod).
func ValidateDeviceMode(mode string) bool {
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
