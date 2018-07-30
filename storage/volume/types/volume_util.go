package types

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/bytefmt"
	"github.com/alibaba/pouch/storage/volume/types/meta"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func translateSelector(k, v string) SelectorRequirement {
	values := strings.Split(v, ",")

	return SelectorRequirement{
		Key:      k,
		Operator: selection.In,
		Values:   values,
	}
}

func buildVolumeConfig(options map[string]string) (*VolumeConfig, error) {
	size := ""
	config := &VolumeConfig{
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

		delete(options, optionSize)
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

// NewVolumeFromID will create an Volume using mountPath, size and VolumeID.
func NewVolumeFromID(mountPath, size string, id VolumeID) *Volume {
	if id.Options == nil {
		id.Options = map[string]string{}
	}
	if id.Labels == nil {
		id.Labels = map[string]string{}
	}
	if id.Selectors == nil {
		id.Selectors = map[string]string{}
	}

	now := time.Now()
	v := &Volume{
		ObjectMeta: meta.ObjectMeta{
			Name:              id.Name,
			Claimer:           "pouch",
			Namespace:         "pouch",
			UID:               uuid.NewRandom().String(),
			Generation:        meta.ObjectPhasePreCreate,
			Labels:            id.Labels,
			CreationTimestamp: &now,
			ModifyTimestamp:   &now,
		},
		Spec: &VolumeSpec{
			Backend:  id.Driver,
			Extra:    id.Options,
			Selector: make(Selector, 0),
			VolumeConfig: &VolumeConfig{
				Size: size,
			},
		},
		Status: &VolumeStatus{
			MountPoint: mountPath,
		},
	}

	for n, selector := range id.Selectors {
		requirement := translateSelector(n, strings.ToLower(selector))
		v.Spec.Selector = append(v.Spec.Selector, requirement)
	}

	return v
}

// NewVolume generates a volume based VolumeID
func NewVolume(id VolumeID) (*Volume, error) {
	now := time.Now()
	v := &Volume{
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
		Spec: &VolumeSpec{
			Extra:    map[string]string{},
			Selector: make(Selector, 0),
		},
		Status: &VolumeStatus{},
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
		v.Spec.Backend = DefaultBackend
		v.Labels["backend"] = DefaultBackend
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

// extractOptionsFromVolumeConfig will extract options from VolumeConfig.
func extractOptionsFromVolumeConfig(config *VolumeConfig) map[string]string {
	var options = map[string]string{}

	if config == nil {
		return options
	}

	if config.Size != "" {
		options[optionSize] = config.Size
	}

	if config.FileSystem != "" {
		options[optionFS] = config.FileSystem
	}

	if config.WriteBPS != 0 {
		options[optionWBps] = strconv.FormatInt(config.WriteBPS, 10)
	}

	if config.ReadBPS != 0 {
		options[optionRBps] = strconv.FormatInt(config.ReadBPS, 10)
	}

	if config.TotalIOPS != 0 {
		options[optionIOps] = strconv.FormatInt(config.TotalIOPS, 10)
	}

	if config.WriteIOPS != 0 {
		options[optionWriteIOps] = strconv.FormatInt(config.WriteIOPS, 10)
	}

	if config.ReadIOPS != 0 {
		options[optionReadIOps] = strconv.FormatInt(config.ReadIOPS, 10)
	}

	return options
}

// ExtractOptionsFromVolume extracts options from a volume.
func ExtractOptionsFromVolume(v *Volume) map[string]string {
	var options map[string]string

	// extract options from volume config.
	options = extractOptionsFromVolumeConfig(v.Spec.VolumeConfig)

	// extract options from selector.
	for _, s := range v.Spec.Selector {
		k := fmt.Sprintf("selector.%s", s.Key)
		options[k] = strings.Join(s.Values, ",")
	}

	// extract options from Extra.
	for k, v := range v.Spec.Extra {
		options[k] = v
	}

	return options
}
