package volume

import (
	"fmt"
	"path"

	"github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/storage/volume/types"

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
