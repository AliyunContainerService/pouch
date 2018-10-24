package types

import (
	"time"

	"github.com/alibaba/pouch/storage/volume/types/meta"

	"github.com/pborman/uuid"
)

// NewVolumeFromContext will create an Volume using mountPath, size and VolumeContext.
func NewVolumeFromContext(mountPath, size string, id VolumeContext) *Volume {
	if id.Options == nil {
		id.Options = map[string]string{}
	}
	if id.Labels == nil {
		id.Labels = map[string]string{}
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
			Backend: id.Driver,
			Extra:   id.Options,
			VolumeConfig: &VolumeConfig{
				Size: size,
			},
		},
		Status: &VolumeStatus{
			MountPoint: mountPath,
		},
	}

	return v
}
