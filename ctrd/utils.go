package ctrd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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
		Proxy: proxyFunc,
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

// generateID generates image's ID by the SHA256 hash of its configuration JSON.
func generateID(config *types.ImageInfo) (digest.Digest, error) {
	var ID digest.Digest

	b, err := json.Marshal(config)
	if err != nil {
		return ID, err
	}

	ID = digest.FromBytes(b)
	return ID, nil
}

// rootFSToAPIType transfer the rootfs from OCI format to Pouch format.
func rootFSToAPIType(rootFs *v1.RootFS) types.ImageInfoRootFS {
	var layers []string
	for _, l := range rootFs.DiffIDs {
		layers = append(layers, l.String())
	}
	return types.ImageInfoRootFS{
		Type:   rootFs.Type,
		Layers: layers,
	}
}

// ociImageToPouchImage transfer the image from OCI format to Pouch format.
func ociImageToPouchImage(ociImage v1.Image) (types.ImageInfo, error) {
	imageConfig := ociImage.Config

	volumes := make(map[string]interface{})
	for k, obj := range imageConfig.Volumes {
		volumes[k] = obj
	}
	cfg := &types.ContainerConfig{
		// TODO: add more fields
		User:       imageConfig.User,
		Env:        imageConfig.Env,
		Entrypoint: imageConfig.Entrypoint,
		Cmd:        imageConfig.Cmd,
		WorkingDir: imageConfig.WorkingDir,
		Labels:     imageConfig.Labels,
		StopSignal: imageConfig.StopSignal,
		Volumes:    volumes,
	}

	rootFs := rootFSToAPIType(&ociImage.RootFS)

	// FIXME need to refactor it and the ociImage's list interface.
	imageInfo := types.ImageInfo{
		Architecture: ociImage.Architecture,
		Config:       cfg,
		CreatedAt:    ociImage.Created.Format(utils.TimeLayout),
		Os:           ociImage.OS,
		RootFS:       &rootFs,
	}
	return imageInfo, nil
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
	var swappiness uint64
	if resources.MemorySwappiness != nil {
		swappiness = uint64(*(resources.MemorySwappiness))
	}
	r.Memory = &specs.LinuxMemory{
		Limit:       &resources.Memory,
		Swap:        &resources.MemorySwap,
		Swappiness:  &swappiness,
		Reservation: &resources.MemoryReservation,
	}

	// TODO: add more fields.

	return r, nil
}
