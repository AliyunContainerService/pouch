package volume

import (
	"fmt"
	"path"
	"strings"

	"github.com/alibaba/pouch/storage/controlserver/client"
	"github.com/alibaba/pouch/storage/volume/driver"
	volerr "github.com/alibaba/pouch/storage/volume/error"
	"github.com/alibaba/pouch/storage/volume/types"
	"github.com/alibaba/pouch/storage/volume/types/meta"

	"github.com/pkg/errors"
)

func (c *Core) volumePath(v *types.Volume, dv driver.Driver) (string, error) {
	p, err := dv.Path(driver.Contexts(), v)
	if err != nil {
		return "", err
	}
	if !path.IsAbs(p) {
		return "", errors.Errorf("Volume path: %s not absolute", p)
	}

	return p, nil
}

func (c *Core) getStorage(id types.StorageID) (*types.Storage, error) {
	if !c.EnableControl {
		return nil, fmt.Errorf("disable control server")
	}

	s := &types.Storage{
		ObjectMeta: meta.ObjectMeta{
			UID: id.UID,
		},
	}

	url, err := c.storageURL(id)
	if err != nil {
		return nil, err
	}
	if err := client.New().Get(url, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Core) storageURL(id ...types.StorageID) (string, error) {
	if c.BaseURL == "" {
		return "", volerr.ErrDisableControl
	}
	if len(id) == 0 {
		return client.JoinURL(c.BaseURL, types.APIVersion, client.StoragePath)
	}
	return client.JoinURL(c.BaseURL, types.APIVersion, client.StoragePath, id[0].UID)
}

func (c *Core) volumeURL(id ...types.VolumeID) (string, error) {
	if c.BaseURL == "" {
		return "", volerr.ErrDisableControl
	}
	if len(id) == 0 {
		return client.JoinURL(c.BaseURL, types.APIVersion, client.VolumePath)
	}
	return client.JoinURL(c.BaseURL, types.APIVersion, client.VolumePath, id[0].Name)
}

func (c *Core) listVolumeURL(labels map[string]string) (string, error) {
	if c.BaseURL == "" {
		return "", volerr.ErrDisableControl
	}
	url, err := client.JoinURL(c.BaseURL, types.APIVersion, client.VolumePath)
	if err != nil {
		return "", err
	}

	querys := make([]string, 0, len(labels))
	for k, v := range labels {
		querys = append(querys, fmt.Sprintf("labels=%s=%s", k, v))
	}

	url = url + "?" + strings.Join(querys, "&")
	return url, nil
}

func (c *Core) listVolumeNameURL(labels map[string]string) (string, error) {
	if c.BaseURL == "" {
		return "", volerr.ErrDisableControl
	}
	url, err := client.JoinURL(c.BaseURL, types.APIVersion, "/listkeys", client.VolumePath)
	if err != nil {
		return "", err
	}

	querys := make([]string, 0, len(labels))
	for k, v := range labels {
		querys = append(querys, fmt.Sprintf("labels=%s=%s", k, v))
	}

	url = url + "?" + strings.Join(querys, "&")
	return url, nil
}

func checkVolume(v *types.Volume) error {
	if v.Spec.ClusterID == "" || v.Status.Phase == types.VolumePhaseFailed {
		err := fmt.Errorf("volume is created failed: %s", v.Name)
		return err
	} else if v.Status.Phase != types.VolumePhaseReady {
		err := fmt.Errorf("volume is being creating: %s", v.Name)
		return err
	}

	return nil
}

func checkOptions(v *types.Volume) error {
	var (
		driverOpts map[string]types.Option
	)

	dv, err := driver.Get(v.Spec.Backend)
	if err != nil {
		return errors.Errorf("failed to get backend driver %s: %v", v.Spec.Backend, err)
	}

	if opt, ok := dv.(driver.Opt); ok {
		driverOpts = opt.Options()
	}

	if driverOpts != nil {
		for name, opt := range driverOpts {
			if _, ok := v.Spec.Extra[name]; !ok {
				v.Spec.Extra[name] = opt.Value
			}
		}
	}

	return nil
}
