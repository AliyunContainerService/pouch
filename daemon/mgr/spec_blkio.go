package mgr

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/alibaba/pouch/apis/types"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupBlkio(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s
	r := meta.HostConfig.Resources

	weightDevice, err := getWeightDevice(r.BlkioWeightDevice)
	if err != nil {
		return err
	}
	readBpsDevice, err := getThrottleDevice(r.BlkioDeviceReadBps)
	if err != nil {
		return err
	}
	writeBpsDevice, err := getThrottleDevice(r.BlkioDeviceWriteBps)
	if err != nil {
		return err
	}
	readIOpsDevice, err := getThrottleDevice(r.BlkioDeviceReadIOps)
	if err != nil {
		return err
	}
	writeIOpsDevice, err := getThrottleDevice(r.BlkioDeviceWriteIOps)
	if err != nil {
		return err
	}

	if s.Linux.Resources == nil {
		s.Linux.Resources = &specs.LinuxResources{}
	}

	s.Linux.Resources.BlockIO = &specs.LinuxBlockIO{
		Weight:                  &r.BlkioWeight,
		WeightDevice:            weightDevice,
		ThrottleReadBpsDevice:   readBpsDevice,
		ThrottleReadIOPSDevice:  readIOpsDevice,
		ThrottleWriteBpsDevice:  writeBpsDevice,
		ThrottleWriteIOPSDevice: writeIOpsDevice,
	}

	return nil
}

func getWeightDevice(devs []*types.WeightDevice) ([]specs.LinuxWeightDevice, error) {
	var stat syscall.Stat_t
	var weightDevice []specs.LinuxWeightDevice

	for _, dev := range devs {
		if err := syscall.Stat(dev.Path, &stat); err != nil {
			return nil, err
		}

		d := specs.LinuxWeightDevice{
			Weight: &dev.Weight,
		}
		d.Major = int64(stat.Rdev / 256)
		d.Minor = int64(stat.Rdev % 256)
		weightDevice = append(weightDevice, d)
	}

	return weightDevice, nil
}

func getThrottleDevice(devs []*types.ThrottleDevice) ([]specs.LinuxThrottleDevice, error) {
	var stat syscall.Stat_t
	var ThrottleDevice []specs.LinuxThrottleDevice

	for _, dev := range devs {
		if err := syscall.Stat(dev.Path, &stat); err != nil {
			return nil, err
		}

		d := specs.LinuxThrottleDevice{
			Rate: dev.Rate,
		}
		d.Major = int64(stat.Rdev / 256)
		d.Minor = int64(stat.Rdev % 256)
		ThrottleDevice = append(ThrottleDevice, d)
	}

	return ThrottleDevice, nil
}

func setupDiskQuota(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	s := spec.s

	rootFSQuota, ok := meta.Config.DiskQuota["/"]
	if !ok || rootFSQuota == "" {
		commonQuota, ok := meta.Config.DiskQuota[".*"]
		if !ok || commonQuota == "" {
			return nil
		}
		rootFSQuota = commonQuota
	}

	if s.Hooks == nil {
		s.Hooks = &specs.Hooks{}
	}
	if s.Hooks.Prestart == nil {
		s.Hooks.Prestart = []specs.Hook{}
	}

	target, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(os.Getpid()), "exe"))
	if err != nil {
		return err
	}

	quotaPrestart := specs.Hook{
		Path: target,
		Args: []string{"set-diskquota", meta.BaseFS, rootFSQuota, meta.Config.QuotaID},
	}
	s.Hooks.Prestart = append(s.Hooks.Prestart, quotaPrestart)

	return nil
}
