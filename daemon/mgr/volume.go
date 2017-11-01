package mgr

import (
	"context"
	"os"
	"path"
	"strings"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/spec"
	"github.com/alibaba/pouch/volume"
	"github.com/alibaba/pouch/volume/types"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// VolumeMgr defines interface to manage container volume.
type VolumeMgr interface {
	// Create is used to create volume.
	Create(ctx context.Context, name, driver string, options, labels map[string]string) error

	// Remove is used to delete an existing volume.
	Remove(ctx context.Context, name string) error

	// List returns all volumes on this host.
	List(ctx context.Context, labels map[string]string) ([]string, error)

	// Info returns the information of volume that specified name/id.
	Info(ctx context.Context, name string) (*types.Volume, error)

	// Attach is used to bind a volume to container.
	Attach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)

	// Detach is used to unbind a volume from container.
	Detach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)
}

// VolumeManager is the default implement of interface VolumeMgr.
type VolumeManager struct {
	core  *volume.Core
	store *meta.Store
}

// NewVolumeManager creates a brand new volume manager.
func NewVolumeManager(ms *meta.Store, cfg types.Config) (*VolumeManager, error) {
	// init voluem config
	cfg.RemoveVolume = true
	cfg.VolumeMetaPath = path.Join(path.Dir(ms.BaseDir), "volume", "volume.db")
	cfg.DefaultBackend = "local"

	core, err := volume.NewCore(cfg)
	if err != nil {
		return nil, err
	}

	// registe setup volume spec functions.
	spec.RegisteSetupFunc(setupMounts)

	return &VolumeManager{
		core:  core,
		store: ms,
	}, nil
}

// Create is used to create volume.
func (vm *VolumeManager) Create(ctx context.Context, name, driver string, options, labels map[string]string) error {
	id := types.VolumeID{
		Name:      name,
		Driver:    driver,
		Options:   map[string]string{},
		Selectors: map[string]string{},
	}

	if labels != nil {
		id.Labels = labels
	} else {
		id.Labels = map[string]string{}
	}

	for key, opt := range options {
		if strings.HasPrefix(key, "selector.") {
			key = strings.TrimPrefix(key, "selector.")
			id.Selectors[key] = opt
			continue
		}

		id.Options[key] = opt
	}

	return vm.core.CreateVolume(id)
}

// Remove is used to delete an existing volume.
func (vm *VolumeManager) Remove(ctx context.Context, name string) error {
	id := types.VolumeID{
		Name: name,
	}
	return vm.core.RemoveVolume(id)
}

// List returns all volumes on this host.
func (vm *VolumeManager) List(ctx context.Context, labels map[string]string) ([]string, error) {
	if _, ok := labels["hostname"]; !ok {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		labels["hostname"] = hostname
	}

	return vm.core.ListVolumeName(labels)
}

// Info returns the information of volume that specified name/id.
func (vm *VolumeManager) Info(ctx context.Context, name string) (*types.Volume, error) {
	id := types.VolumeID{
		Name: name,
	}
	return vm.core.GetVolume(id)
}

// Attach is used to bind a volume to container.
func (vm *VolumeManager) Attach(ctx context.Context, name string, options map[string]string) (*types.Volume, error) {
	id := types.VolumeID{
		Name: name,
	}
	return vm.core.AttachVolume(id, options)
}

// Detach is used to unbind a volume from container.
func (vm *VolumeManager) Detach(ctx context.Context, name string, options map[string]string) (*types.Volume, error) {
	id := types.VolumeID{
		Name: name,
	}
	return vm.core.DetachVolume(id, options)
}

func setupMounts(ctx context.Context, c *apitypes.ContainerInfo, s *specs.Spec) error {
	mounts := s.Mounts
	if c.Config.HostConfig == nil {
		return nil
	}
	for _, v := range c.Config.HostConfig.Binds {
		sd := strings.SplitN(v, ":", 2)
		if len(sd) != 2 {
			continue
		}
		mounts = append(mounts, specs.Mount{
			Destination: sd[0],
			Source:      sd[1],
			Type:        "bind",
			Options:     []string{"rbind"},
		})
	}
	s.Mounts = mounts
	return nil
}
