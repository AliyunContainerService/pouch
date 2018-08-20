package volume

import (
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
