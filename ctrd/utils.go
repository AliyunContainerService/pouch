package ctrd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/runtime/linux/runctypes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func withExitShimV1CheckpointTaskOpts() containerd.CheckpointTaskOpts {
	return func(r *containerd.CheckpointTaskInfo) error {
		r.Options = &runctypes.CheckpointOptions{
			Exit: true,
		}
		return nil
	}
}

// isInsecureDomain will return true if the domain of reference is in the
// insecure registry. The insecure registry will accept HTTP or HTTPS with
// certificates from unknown CAs.
func (c *Client) isInsecureDomain(ref string) bool {
	u, err := url.Parse("dummy://" + ref)
	if err != nil {
		logrus.Warningf("failed to parse reference(%s) into url: %v", ref, err)
		return false
	}

	for _, r := range c.insecureRegistries {
		if r == u.Host {
			return true
		}
	}
	return false
}

// resolverWrapper wrap a image resolver
// do reference <-> name translation before each operation.
type resolverWrapper struct {
	refToName map[string]string
	resolver  remotes.Resolver
}

// Resolve attempts to resolve the reference into a name and descriptor.
// translate the reference to a name which may be a short name like 'library/ubuntu'.
func (r *resolverWrapper) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	newRef, desc, err := r.resolver.Resolve(ctx, ref)
	if err != nil {
		return "", ocispec.Descriptor{}, err
	}

	if name, ok := r.refToName[newRef]; ok {
		return name, desc, nil
	}

	return newRef, desc, nil
}

// Fetcher returns a new fetcher for the provided reference.
// All content fetched from the returned fetcher will be
// from the namespace referred to by ref.
func (r *resolverWrapper) Fetcher(ctx context.Context, name string) (remotes.Fetcher, error) {
	ref := name
	for rf, n := range r.refToName {
		if name == n {
			ref = rf
			break
		}
	}
	return r.resolver.Fetcher(ctx, ref)
}

// Pusher returns a new pusher for the provided reference
func (r *resolverWrapper) Pusher(ctx context.Context, name string) (remotes.Pusher, error) {
	ref := name
	for rf, n := range r.refToName {
		if name == n {
			ref = rf
			break
		}
	}
	return r.resolver.Pusher(ctx, ref)
}

func newImageResolver(refToName map[string]string, resolverOpt docker.ResolverOptions) remotes.Resolver {
	return &resolverWrapper{
		refToName: refToName,
		resolver:  docker.NewResolver(resolverOpt),
	}
}

// getResolver try to resolve ref in the reference list, return the resolver and the first available ref.
func (c *Client) getResolver(ctx context.Context, authConfig *types.AuthConfig, name string, refs []string, resolverOpt docker.ResolverOptions) (remotes.Resolver, string, error) {
	username, secret := "", ""
	if authConfig != nil {
		username = authConfig.Username
		secret = authConfig.Password
	}

	var (
		availableRef string
		opt          docker.ResolverOptions
	)

	for _, ref := range refs {
		namedRef, err := reference.Parse(ref)
		if err != nil {
			logrus.Warnf("failed to parse image reference when trying to resolve image %s, raw reference is %s: %v", ref, name, err)
			continue
		}
		namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))

		insecure := c.isInsecureDomain(ref)
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

		opt = docker.ResolverOptions{
			Tracker:   resolverOpt.Tracker,
			PlainHTTP: insecure,
			Credentials: func(host string) (string, string, error) {
				// Only one host
				return username, secret, nil
			},
			Client: &http.Client{
				Transport: tr,
			},
		}

		resolver := docker.NewResolver(opt)

		if _, _, err := resolver.Resolve(ctx, namedRef.String()); err == nil {
			availableRef = namedRef.String()
			break
		}
	}

	if availableRef == "" {
		return nil, "", fmt.Errorf("there is no available image reference after trying %+q", refs)
	}

	refToName := map[string]string{
		availableRef: name,
	}

	return newImageResolver(refToName, opt), availableRef, nil
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
		Kernel:      &resources.KernelMemory,
		// TODO: add other fields of specs.LinuxMemory
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
