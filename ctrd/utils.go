package ctrd

import (
	"crypto/tls"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/runtime/linux/runctypes"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

func withExitShimV1CheckpointTaskOpts() containerd.CheckpointTaskOpts {
	return func(r *containerd.CheckpointTaskInfo) error {
		r.Options = &runctypes.CheckpointOptions{
			Exit: true,
		}
		return nil
	}
}

func resolver(authConfig *types.AuthConfig, resolverOpt docker.ResolverOptions) (remotes.Resolver, error) {
	var (
		// TODO
		username = ""
		secret   = ""
		refresh  = ""
		insecure = false
	)

	if authConfig != nil {
		username = authConfig.Username
		secret = authConfig.Password
	}

	// FIXME
	_ = refresh

	options := docker.ResolverOptions{
		PlainHTTP: resolverOpt.PlainHTTP,
		Tracker:   resolverOpt.Tracker,
	}
	options.Credentials = func(host string) (string, string, error) {
		// Only one host
		return username, secret, nil
	}

	tr := &http.Transport{
		Proxy: proxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		ExpectContinueTimeout: 5 * time.Second,
	}

	options.Client = &http.Client{
		Transport: tr,
	}

	return docker.NewResolver(options), nil
}

// GetWeightDevice Convert weight device from []*types.WeightDevice to []specs.LinuxWeightDevice
func GetWeightDevice(devs []*types.WeightDevice) ([]specs.LinuxWeightDevice, error) {
	var stat syscall.Stat_t
	var weightDevice []specs.LinuxWeightDevice

	for _, dev := range devs {
		if err := syscall.Stat(dev.Path, &stat); err != nil {
			return nil, err
		}

		d := specs.LinuxWeightDevice{
			Weight: &dev.Weight,
		}
		d.Major = int64(stat.Rdev >> 8)
		d.Minor = int64(stat.Rdev & 255)
		weightDevice = append(weightDevice, d)
	}

	return weightDevice, nil
}

// GetThrottleDevice Convert throttle device from []*types.ThrottleDevice to []specs.LinuxThrottleDevice
func GetThrottleDevice(devs []*types.ThrottleDevice) ([]specs.LinuxThrottleDevice, error) {
	var stat syscall.Stat_t
	var ThrottleDevice []specs.LinuxThrottleDevice

	for _, dev := range devs {
		if err := syscall.Stat(dev.Path, &stat); err != nil {
			return nil, err
		}

		d := specs.LinuxThrottleDevice{
			Rate: dev.Rate,
		}
		d.Major = int64(stat.Rdev >> 8)
		d.Minor = int64(stat.Rdev & 255)
		ThrottleDevice = append(ThrottleDevice, d)
	}

	return ThrottleDevice, nil
}

// toLinuxResources transfers Pouch Resources to LinuxResources.
func toLinuxResources(resources types.Resources) (*specs.LinuxResources, error) {
	r := &specs.LinuxResources{}

	// toLinuxBlockIO
	readBpsDevice, err := GetThrottleDevice(resources.BlkioDeviceReadBps)
	if err != nil {
		return nil, err
	}
	readIOpsDevice, err := GetThrottleDevice(resources.BlkioDeviceReadIOps)
	if err != nil {
		return nil, err
	}
	writeBpsDevice, err := GetThrottleDevice(resources.BlkioDeviceWriteBps)
	if err != nil {
		return nil, err
	}
	writeIOpsDevice, err := GetThrottleDevice(resources.BlkioDeviceWriteIOps)
	if err != nil {
		return nil, err
	}
	r.BlockIO = &specs.LinuxBlockIO{
		Weight:                  &resources.BlkioWeight,
		ThrottleReadBpsDevice:   readBpsDevice,
		ThrottleReadIOPSDevice:  readIOpsDevice,
		ThrottleWriteBpsDevice:  writeBpsDevice,
		ThrottleWriteIOPSDevice: writeIOpsDevice,
	}

	// toLinuxCPU
	shares := uint64(resources.CPUShares)
	period := uint64(resources.CPUPeriod)
	r.CPU = &specs.LinuxCPU{
		Cpus:   resources.CpusetCpus,
		Mems:   resources.CpusetMems,
		Shares: &shares,
		Period: &period,
		Quota:  &resources.CPUQuota,
	}

	// toLinuxMemory
	r.Memory = &specs.LinuxMemory{
		Limit:       &resources.Memory,
		Swap:        &resources.MemorySwap,
		Reservation: &resources.MemoryReservation,
	}

	// TODO: add more fields.

	return r, nil
}

// convertCtrdErr converts containerd client error into a pouchd manager error.
// containerd client error converts GRPC code from containerd API to containerd client error.
// pouchd manager error is used in the whole managers and API layers to construct status code for API.
// there should be a way convert the previous to the latter one.
func convertCtrdErr(err error) error {
	if err == nil {
		return nil
	}

	if errdefs.IsNotFound(err) {
		return errors.Wrap(errtypes.ErrNotfound, err.Error())
	}

	if errdefs.IsAlreadyExists(err) {
		return errors.Wrap(errtypes.ErrAlreadyExisted, err.Error())
	}

	if errdefs.IsInvalidArgument(err) {
		return errors.Wrap(errtypes.ErrInvalidParam, err.Error())
	}

	if errdefs.IsNotImplemented(err) {
		return errors.Wrap(errtypes.ErrNotImplemented, err.Error())
	}

	return err
}
