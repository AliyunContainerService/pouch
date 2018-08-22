package types

import (
	"strings"
	"time"

	"github.com/alibaba/pouch/storage/volume/types/meta"

	"github.com/pborman/uuid"
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
