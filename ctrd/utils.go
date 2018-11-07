package ctrd

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

// NewDefaultSpec new a template spec with default.
func NewDefaultSpec(ctx context.Context, id string) (*specs.Spec, error) {
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)
	return oci.GenerateSpec(ctx, nil, &containers.Container{ID: id})
}

func resolver(authConfig *types.AuthConfig) (remotes.Resolver, error) {
	var (
		// TODO
		username  = ""
		secret    = ""
		plainHTTP = false
		refresh   = ""
		insecure  = false
	)

	if authConfig != nil {
		username = authConfig.Username
		secret = authConfig.Password
	}

	// FIXME
	_ = refresh

	options := docker.ResolverOptions{
		PlainHTTP: plainHTTP,
		Tracker:   docker.NewInMemoryTracker(),
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

// toLinuxResources transfers Pouch Resources to LinuxResources.
func toLinuxResources(resources types.Resources) (*specs.LinuxResources, error) {
	r := &specs.LinuxResources{}

	// toLinuxBlockIO
	r.BlockIO = &specs.LinuxBlockIO{
		Weight: &resources.BlkioWeight,
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
