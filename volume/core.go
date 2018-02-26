package volume

import (
	"fmt"

	"github.com/alibaba/pouch/extra/libnetwork/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/alibaba/pouch/pkg/client"
	"github.com/alibaba/pouch/volume/driver"
	volerr "github.com/alibaba/pouch/volume/error"
	"github.com/alibaba/pouch/volume/store"
	"github.com/alibaba/pouch/volume/types"
	"github.com/alibaba/pouch/volume/types/meta"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Core represents volume core struct.
type Core struct {
	Config
	BaseURL       string
	EnableControl bool
}

// NewCore returns Core struct instance with volume config.
func NewCore(cfg Config) (*Core, error) {
	c := &Core{Config: cfg}
	if cfg.ControlAddress != "" {
		c.EnableControl = true
		c.BaseURL = cfg.ControlAddress
	} else {
		c.EnableControl = false
	}

	if err := store.MetaNewStore(cfg.VolumeMetaPath); err != nil {
		return nil, err
	}

	return c, nil
}

// GetVolume return a volume's info with specified name, If not errors.
func (c *Core) GetVolume(id types.VolumeID) (*types.Volume, error) {
	v := &types.Volume{
		ObjectMeta: meta.ObjectMeta{
			Name: id.Name,
		},
	}

	// first, try to get volume from local store.
	err := store.MetaGet(v)
	if err == nil {
		return v, nil
	}
	cerr, ok := err.(volerr.CoreError)
	if !ok {
		return nil, err
	}
	if !cerr.IsLocalMetaNotfound() {
		return nil, err
	}
	err = volerr.ErrVolumeNotfound

	// then, try to get volume from central store.
	if c.EnableControl {
		url, err := c.volumeURL(id)
		if err != nil {
			return nil, err
		}

		if err = client.New().Get(url, v); err == nil {
			return v, nil
		}
		if ce, ok := err.(client.Error); ok && ce.IsNotfound() {
			return nil, volerr.ErrVolumeNotfound
		}
		return nil, err
	}

	return nil, err
}

// ExistVolume return 'true' if volume be found and not errors.
func (c *Core) ExistVolume(id types.VolumeID) (bool, error) {
	_, err := c.GetVolume(id)
	if err != nil {
		if ec, ok := err.(volerr.CoreError); ok && ec.IsVolumeNotfound() {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateVolume use to create a volume, if failed, will return error info.
func (c *Core) CreateVolume(id types.VolumeID) error {
	exist, err := c.ExistVolume(id)
	if err != nil {
		return err
	}
	if exist {
		return volerr.ErrVolumeExisted
	}

	v, err := c.newVolume(id)
	if err != nil {
		return errors.Wrapf(err, "Create volume")
	}

	dv, ok := driver.Get(v.Spec.Backend)
	if !ok {
		return errors.Errorf("Backend driver: %s not found", v.Spec.Backend)
	}

	p, err := c.volumePath(v, dv)
	if err != nil {
		return err
	}
	v.SetPath(p)

	// check options, then delete invalid options.
	if err := checkOptions(v); err != nil {
		return err
	}

	// Create volume's meta.
	ctx := driver.Contexts()

	if !dv.StoreMode(ctx).UseLocalMeta() {
		url, err := c.volumeURL()
		if err != nil {
			return err
		}

		if err := client.New().Create(url, v); err != nil {
			return errors.Wrap(err, "Create volume")
		}
	}

	// Create volume's store room on local.
	var s *types.Storage
	if !dv.StoreMode(ctx).IsLocal() {
		s, err = c.getStorage(v.StorageID())
		if err != nil {
			return err
		}
	}

	if !dv.StoreMode(ctx).CentralCreateDelete() {
		if err := dv.Create(ctx, v, s); err != nil {
			return err
		}

		if err := store.MetaPut(v); err != nil {
			return err
		}
	}

	if f, ok := dv.(driver.Formator); ok {
		err := f.Format(ctx, v, s)
		if err == nil {
			return nil
		}

		logrus.Errorf("failed to format new volume: %s, err: %v", v.Name, err)
		logrus.Warnf("rollback create volume, start to cleanup new volume: %s", v.Name)
		if err := c.RemoveVolume(id); err != nil {
			logrus.Errorf("failed to rollback create volume, cleanup new volume: %s, err: %v", v.Name, err)
			return err
		}

		// return format error.
		return err
	}

	return nil
}

// ListVolumeName return the name of all volumes only.
// Param 'labels' use to filter the volume's names, only return those you want.
func (c *Core) ListVolumeName(labels map[string]string) ([]string, error) {
	var names []string

	// first, list local meta store.
	volumes, err := store.MetaList()
	if err != nil {
		return nil, err
	}

	// then, list central store.
	if c.EnableControl {
		url, err := c.listVolumeNameURL(labels)
		if err != nil {
			return nil, errors.Wrap(err, "List volume's name")
		}

		log.Debugf("List volume URL: %s, labels: %s", url, labels)

		if err := client.New().ListKeys(url, &names); err != nil {
			return nil, errors.Wrap(err, "List volume's name")
		}
	}

	for _, v := range volumes {
		names = append(names, v.GetName())
	}

	return names, nil
}

// RemoveVolume remove volume from storage and meta information, if not success return error.
func (c *Core) RemoveVolume(id types.VolumeID) error {
	v, dv, err := c.GetVolumeDriver(id)
	if err != nil {
		return errors.Wrap(err, "Remove volume: "+id.String())
	}

	// Call interface to remove meta info.
	if dv.StoreMode(driver.Contexts()).UseLocalMeta() {
		if err := store.MetaDel(id.Name); err != nil {
			return err
		}
	} else {
		url, err := c.volumeURL(id)
		if err != nil {
			return errors.Wrap(err, "Remove volume: "+id.String())
		}
		if err := client.New().Delete(url, v); err != nil {
			return errors.Wrap(err, "Remove volume: "+id.String())
		}
	}

	// Call driver's Remove method to remove the volume.
	if !dv.StoreMode(driver.Contexts()).CentralCreateDelete() {
		var s *types.Storage
		if !dv.StoreMode(driver.Contexts()).UseLocalMeta() {
			s, err = c.getStorage(v.StorageID())
			if err != nil {
				return err
			}
		}

		if err := dv.Remove(driver.Contexts(), v, s); err != nil {
			return err
		}
	}

	return nil
}

// VolumePath return the path of volume on node host.
func (c *Core) VolumePath(id types.VolumeID) (string, error) {
	v, dv, err := c.GetVolumeDriver(id)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Get volume: %s path", id.String()))
	}

	return c.volumePath(v, dv)
}

// GetVolumeDriver return the backend driver and volume with specified volume's id.
func (c *Core) GetVolumeDriver(id types.VolumeID) (*types.Volume, driver.Driver, error) {
	v, err := c.GetVolume(id)
	if err != nil {
		return nil, nil, err
	}
	dv, ok := driver.Get(v.Spec.Backend)
	if !ok {
		return nil, nil, errors.Errorf("Backend driver: %s not found", v.Spec.Backend)
	}
	return v, dv, nil
}

// AttachVolume to enable a volume on local host.
func (c *Core) AttachVolume(id types.VolumeID, extra map[string]string) (*types.Volume, error) {
	v, dv, err := c.GetVolumeDriver(id)
	if err != nil {
		return nil, err
	}

	ctx := driver.Contexts()
	var s *types.Storage

	// merge extra to volume spec extra.
	for key, value := range extra {
		v.Spec.Extra[key] = value
	}

	if a, ok := dv.(driver.AttachDetach); ok {
		if !dv.StoreMode(ctx).IsLocal() {
			if s, err = c.getStorage(v.StorageID()); err != nil {
				return nil, err
			}
		}

		if err = a.Attach(ctx, v, s); err != nil {
			return nil, err
		}
	}

	// Call interface to update meta info.
	if dv.StoreMode(driver.Contexts()).UseLocalMeta() {
		if err := store.MetaPut(v); err != nil {
			return nil, err
		}
	} else {
		url, err := c.volumeURL(id)
		if err != nil {
			return nil, errors.Wrap(err, "Update volume: "+id.String())
		}
		if err := client.New().Update(url, v); err != nil {
			return nil, errors.Wrap(err, "Update volume: "+id.String())
		}
	}

	return v, nil
}

// DetachVolume to disable a volume on local host.
func (c *Core) DetachVolume(id types.VolumeID, extra map[string]string) (*types.Volume, error) {
	v, dv, err := c.GetVolumeDriver(id)
	if err != nil {
		return nil, err
	}

	ctx := driver.Contexts()
	var s *types.Storage

	// merge extra to volume spec extra.
	for key, value := range extra {
		v.Spec.Extra[key] = value
	}

	if a, ok := dv.(driver.AttachDetach); ok {
		if !dv.StoreMode(ctx).IsLocal() {
			if s, err = c.getStorage(v.StorageID()); err != nil {
				return nil, err
			}
		}

		if err = a.Detach(ctx, v, s); err != nil {
			return nil, err
		}
	}

	// Call interface to update meta info.
	if dv.StoreMode(driver.Contexts()).UseLocalMeta() {
		if err := store.MetaPut(v); err != nil {
			return nil, err
		}
	} else {
		url, err := c.volumeURL(id)
		if err != nil {
			return nil, errors.Wrap(err, "Update volume: "+id.String())
		}
		if err := client.New().Update(url, v); err != nil {
			return nil, errors.Wrap(err, "Update volume: "+id.String())
		}
	}

	return v, nil
}
