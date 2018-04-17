package volume

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/storage/controlserver/client"
	"github.com/alibaba/pouch/storage/volume/driver"
	volerr "github.com/alibaba/pouch/storage/volume/error"
	"github.com/alibaba/pouch/storage/volume/types"
	"github.com/alibaba/pouch/storage/volume/types/meta"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

func (c *Core) newVolume(id types.VolumeID) (*types.Volume, error) {
	now := time.Now()
	v := &types.Volume{
		ObjectMeta: meta.ObjectMeta{
			Name:              id.Name,
			Claimer:           "pouch",
			Namespace:         "pouch",
			UID:               uuid.NewRandom().String(),
			Generation:        meta.ObjectPhasePreCreate,
			Labels:            labels.Set{},
			CreationTimestamp: &now,
			ModifyTimestamp:   &now,
		},
		Spec: &types.VolumeSpec{
			Extra:    map[string]string{},
			Selector: make(types.Selector, 0),
		},
		Status: &types.VolumeStatus{},
	}

	conf, err := buildVolumeConfig(id.Options)
	if err != nil {
		return nil, err
	}
	v.Spec.VolumeConfig = conf

	for n, opt := range id.Options {
		v.Spec.Extra[n] = opt
	}

	for n, selector := range id.Selectors {
		requirement := translateSelector(n, strings.ToLower(selector))
		v.Spec.Selector = append(v.Spec.Selector, requirement)
	}

	v.Labels = id.Labels

	// initialize default option/label/selector
	if id.Driver != "" {
		v.Spec.Backend = id.Driver
		v.Labels["backend"] = id.Driver
	} else {
		v.Spec.Backend = c.DefaultBackend
		v.Labels["backend"] = c.DefaultBackend
	}

	if hostname, err := os.Hostname(); err == nil {
		v.Labels["hostname"] = hostname
	}

	if _, ok := id.Selectors[selectNamespace]; !ok {
		requirement := translateSelector("namespace", commonOptions["namespace"].Value)
		v.Spec.Selector = append(v.Spec.Selector, requirement)
	}

	if _, ok := v.Spec.Extra["sifter"]; !ok {
		v.Spec.Extra["sifter"] = "Default"
	}

	return v, nil
}

func translateSelector(k, v string) types.SelectorRequirement {
	values := strings.Split(v, ",")

	return types.SelectorRequirement{
		Key:      k,
		Operator: selection.In,
		Values:   values,
	}
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

func buildVolumeConfig(options map[string]string) (*types.VolumeConfig, error) {
	size := ""
	config := &types.VolumeConfig{
		FileSystem: defaultFileSystem,
		MountOpt:   defaultFileSystem,
	}

	// Parse size
	if s, ok := options[optionSize]; ok {
		size = s
	}

	if size != "" {
		sizeInt, err := bytefmt.ToMegabytes(size)
		if err != nil {
			return nil, err
		}
		config.Size = strconv.Itoa(int(sizeInt)) + "M"
	}

	// Parse filesystem
	if fs, ok := options[optionFS]; ok {
		config.FileSystem = fs
		delete(options, optionFS)
	}
	config.MountOpt = strings.Split(config.FileSystem, " ")[0]

	// Parse IO config
	if wbps, ok := options[optionWBps]; ok {
		v, err := strconv.ParseInt(wbps, 10, 64)
		if err != nil {
			return nil, err
		}
		config.WriteBPS = v

		delete(options, optionWBps)
	}

	if rbps, ok := options[optionRBps]; ok {
		v, err := strconv.ParseInt(rbps, 10, 64)
		if err != nil {
			return nil, err
		}
		config.ReadBPS = v

		delete(options, optionRBps)
	}

	if iops, ok := options[optionIOps]; ok {
		v, err := strconv.ParseInt(iops, 10, 64)
		if err != nil {
			return nil, err
		}
		config.TotalIOPS = v
		delete(options, optionIOps)
	}

	if wiops, ok := options[optionWriteIOps]; ok {
		v, err := strconv.ParseInt(wiops, 10, 64)
		if err != nil {
			return nil, err
		}
		config.WriteIOPS = v
		delete(options, optionWriteIOps)
	}

	if riops, ok := options[optionReadIOps]; ok {
		v, err := strconv.ParseInt(riops, 10, 64)
		if err != nil {
			return nil, err
		}
		config.ReadIOPS = v
		delete(options, optionReadIOps)
	}

	return config, nil
}

func checkOptions(v *types.Volume) error {
	var (
		deleteOpts []string
		driverOpts map[string]types.Option
	)

	dv, ok := driver.Get(v.Spec.Backend)
	if !ok {
		return errors.Errorf("Backend driver: %s not found", v.Spec.Backend)
	}

	if opt, ok := dv.(driver.Opt); ok {
		driverOpts = opt.Options()
	}

	// check extra options is invalid or not.
	for name := range v.Spec.Extra {
		if _, ok := commonOptions[name]; ok {
			continue
		}
		if driverOpts != nil {
			if _, ok := driverOpts[name]; ok {
				continue
			}
		}
		deleteOpts = append(deleteOpts, name)
	}
	for _, d := range deleteOpts {
		delete(v.Spec.Extra, d)
	}

	// set driver options into extra map.
	if driverOpts != nil {
		for name, opt := range driverOpts {
			if _, ok := v.Spec.Extra[name]; !ok {
				v.Spec.Extra[name] = opt.Value
			}
		}
	}

	return nil
}
