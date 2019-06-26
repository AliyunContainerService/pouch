package mgr

import (
	"context"
	"strings"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/daemon/events"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/storage/volume"
	"github.com/alibaba/pouch/storage/volume/types"

	"github.com/pkg/errors"
)

// the filter tags set allowed when pouch volume ls -f
var acceptedVolumeFilterTags = map[string]bool{
	"driver": true,
	"name":   true,
	"label":  true,
}

// VolumeMgr defines interface to manage container volume.
type VolumeMgr interface {
	// Create is used to create volume.
	Create(ctx context.Context, name, driver string, options, labels map[string]string) (*types.Volume, error)

	// Get returns the information of volume that specified name/id.
	Get(ctx context.Context, name string) (*types.Volume, error)

	// List returns all volumes on this host.
	List(ctx context.Context, filter filters.Args) ([]*types.Volume, error)

	// Remove is used to delete an existing volume.
	Remove(ctx context.Context, name string) error

	// Path returns the mount path of volume.
	Path(ctx context.Context, name string) (string, error)

	// Attach is used to bind a volume to container.
	Attach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)

	// Detach is used to unbind a volume from container.
	Detach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)
}

// VolumeManager is the default implement of interface VolumeMgr.
type VolumeManager struct {
	core          *volume.Core
	eventsService *events.Events
}

// NewVolumeManager creates a brand new volume manager.
func NewVolumeManager(cfg volume.Config, eventsService *events.Events) (*VolumeManager, error) {
	// init volume config
	cfg.RemoveVolume = true
	cfg.DefaultBackend = types.DefaultBackend

	core, err := volume.NewCore(cfg)
	if err != nil {
		return nil, err
	}

	return &VolumeManager{
		core:          core,
		eventsService: eventsService,
	}, nil
}

// Create is used to create volume.
func (vm *VolumeManager) Create(ctx context.Context, name, driver string, options, labels map[string]string) (*types.Volume, error) {
	if driver == "" {
		driver = types.DefaultBackend
	}

	id := types.VolumeContext{
		Name:    name,
		Driver:  driver,
		Options: map[string]string{},
		Labels:  map[string]string{},
	}

	if labels != nil {
		id.Labels = labels
	}

	if options != nil {
		id.Options = options
	}

	v, err := vm.core.CreateVolume(id)
	if err != nil {
		if errtypes.IsVolumeExisted(err) {
			return v, nil
		}
		return nil, err
	}

	vm.LogVolumeEvent(ctx, name, "create", map[string]string{"driver": driver})

	return v, nil
}

// Get returns the information of volume that specified name/id.
func (vm *VolumeManager) Get(ctx context.Context, name string) (*types.Volume, error) {
	id := types.VolumeContext{
		Name: name,
	}
	vol, err := vm.core.GetVolume(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.Wrap(errtypes.ErrVolumeNotFound, err.Error())
		}
		return nil, err
	}
	return vol, nil
}

// List returns all volumes on this host.
func (vm *VolumeManager) List(ctx context.Context, filter filters.Args) ([]*types.Volume, error) {
	if err := filter.Validate(acceptedVolumeFilterTags); err != nil {
		return nil, err
	}
	return vm.core.ListVolumes(filter)
}

// Remove is used to delete an existing volume.
func (vm *VolumeManager) Remove(ctx context.Context, name string) error {
	vol, err := vm.Get(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "failed to get volume(%s)", name)
	}

	ref := vol.Option(types.OptionRef)
	if ref != "" {
		return errors.Wrapf(errtypes.ErrVolumeInUse, "failed to remove volume(%s)", name)
	}

	id := types.VolumeContext{
		Name: name,
	}
	if err := vm.core.RemoveVolume(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.Wrap(errtypes.ErrVolumeNotFound, err.Error())
		}
		return err
	}

	vm.LogVolumeEvent(ctx, name, "destroy", map[string]string{"driver": vol.Driver()})

	return nil
}

// Path returns the mount path of volume.
func (vm *VolumeManager) Path(ctx context.Context, name string) (string, error) {
	id := types.VolumeContext{
		Name: name,
	}
	return vm.core.VolumePath(id)
}

// Attach is used to bind a volume to container.
func (vm *VolumeManager) Attach(ctx context.Context, name string, options map[string]string) (*types.Volume, error) {
	id := types.VolumeContext{
		Name: name,
	}

	v, err := vm.Get(ctx, name)
	if err != nil {
		return nil, errors.Errorf("failed to get volume(%s): %v", name, err)
	}

	if options == nil {
		options = make(map[string]string)
	}

	cid, ok := options[types.OptionRef]
	if ok && cid != "" {
		ref := v.Option(types.OptionRef)
		if ref == "" {
			options[types.OptionRef] = cid
		} else if !strings.Contains(ref, cid) {
			options[types.OptionRef] = strings.Join([]string{ref, cid}, ",")
		}
	}

	vm.LogVolumeEvent(ctx, name, "attach", map[string]string{"driver": v.Driver()})
	return vm.core.AttachVolume(id, options)
}

// Detach is used to unbind a volume from container.
func (vm *VolumeManager) Detach(ctx context.Context, name string, options map[string]string) (*types.Volume, error) {
	id := types.VolumeContext{
		Name: name,
	}

	v, err := vm.Get(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get volume(%s)", name)
	}

	if options == nil {
		options = make(map[string]string)
	}

	cid, ok := options[types.OptionRef]
	if ok && cid != "" {
		ref := v.Option(types.OptionRef)
		if !strings.Contains(ref, cid) {
			return v, nil
		}

		if ref != "" {
			ids := strings.Split(ref, ",")
			ids = utils.StringSliceDelete(ids, cid)
			if len(ids) > 0 {
				options[types.OptionRef] = strings.Join(ids, ",")
			} else {
				options[types.OptionRef] = ""
			}
		}
	}
	vm.LogVolumeEvent(ctx, name, "detach", map[string]string{"driver": v.Driver()})
	return vm.core.DetachVolume(id, options)
}
