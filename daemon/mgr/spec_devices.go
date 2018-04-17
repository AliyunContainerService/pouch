package mgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/apis/opts"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/devices"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func u32Ptr(i int64) *uint32     { u := uint32(i); return &u }
func fmPtr(i int64) *os.FileMode { fm := os.FileMode(i); return &fm }

// linuxDevice convert a libcontainer configs.Device to a specs.LinuxDevice object.
func linuxDevice(d *configs.Device) specs.LinuxDevice {
	return specs.LinuxDevice{
		Type:     string(d.Type),
		Path:     d.Path,
		Major:    d.Major,
		Minor:    d.Minor,
		FileMode: fmPtr(int64(d.FileMode)),
		UID:      u32Ptr(int64(d.Uid)),
		GID:      u32Ptr(int64(d.Gid)),
	}
}

func deviceCgroup(d *configs.Device) specs.LinuxDeviceCgroup {
	t := string(d.Type)
	return specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   t,
		Major:  &d.Major,
		Minor:  &d.Minor,
		Access: d.Permissions,
	}
}

func devicesFromPath(pathOnHost, pathInContainer, cgroupPermissions string) (devs []specs.LinuxDevice, devPermissions []specs.LinuxDeviceCgroup, err error) {
	resolvedPathOnHost := pathOnHost

	// check if it is a symbolic link
	if src, e := os.Lstat(pathOnHost); e == nil && src.Mode()&os.ModeSymlink == os.ModeSymlink {
		if linkedPathOnHost, e := filepath.EvalSymlinks(pathOnHost); e == nil {
			resolvedPathOnHost = linkedPathOnHost
		}
	}

	device, err := devices.DeviceFromPath(resolvedPathOnHost, cgroupPermissions)
	if err == nil {
		device.Path = pathInContainer
		return append(devs, linuxDevice(device)), append(devPermissions, deviceCgroup(device)), nil
	}

	// if the device is not a device node
	// try to see if it's a directory holding many devices
	if err == devices.ErrNotADevice {

		// check if it is a directory
		if src, e := os.Stat(resolvedPathOnHost); e == nil && src.IsDir() {

			// mount the internal devices recursively
			filepath.Walk(resolvedPathOnHost, func(dpath string, f os.FileInfo, e error) error {
				childDevice, e := devices.DeviceFromPath(dpath, cgroupPermissions)
				if e != nil {
					// ignore the device
					return nil
				}

				// add the device to userSpecified devices
				childDevice.Path = strings.Replace(dpath, resolvedPathOnHost, pathInContainer, 1)
				devs = append(devs, linuxDevice(childDevice))
				devPermissions = append(devPermissions, deviceCgroup(childDevice))

				return nil
			})
		}
	}

	if len(devs) > 0 {
		return devs, devPermissions, nil
	}

	return devs, devPermissions, fmt.Errorf("error gathering device information while adding custom device %q: %s", pathOnHost, err)
}

func setupDevices(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	var devs []specs.LinuxDevice
	s := spec.s
	devPermissions := s.Linux.Resources.Devices
	if meta.HostConfig.Privileged {
		hostDevices, err := devices.HostDevices()
		if err != nil {
			return err
		}
		for _, d := range hostDevices {
			devs = append(devs, linuxDevice(d))
		}
		devPermissions = []specs.LinuxDeviceCgroup{
			{
				Allow:  true,
				Access: "rwm",
			},
		}
	} else {
		for _, deviceMapping := range meta.HostConfig.Devices {
			if !opts.ValidateDeviceMode(deviceMapping.CgroupPermissions) {
				return fmt.Errorf("%s invalid device mode: %s", deviceMapping.PathOnHost, deviceMapping.CgroupPermissions)
			}
			d, dPermissions, err := devicesFromPath(deviceMapping.PathOnHost, deviceMapping.PathInContainer, deviceMapping.CgroupPermissions)
			if err != nil {
				return err
			}
			devs = append(devs, d...)
			devPermissions = append(devPermissions, dPermissions...)
		}
	}

	if s.Linux.Devices == nil {
		s.Linux.Devices = []specs.LinuxDevice{}
	}
	s.Linux.Devices = append(s.Linux.Devices, devs...)
	s.Linux.Resources.Devices = devPermissions
	return nil
}
