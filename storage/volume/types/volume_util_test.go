package types

import "testing"

func TestNewVolumeFromContext(t *testing.T) {
	tests := []struct {
		Name       string
		Driver     string
		MountPoint string
		Size       string
	}{
		{
			Name:       "volumetest1",
			Driver:     "local",
			MountPoint: "/mnt",
			Size:       "10g",
		},
		{
			Name:       "volumetest2",
			Driver:     "tmpfs",
			MountPoint: "/tmp",
			Size:       "1g",
		},
	}

	for _, tt := range tests {
		volumeID := NewVolumeContext(tt.Name, tt.Driver, nil, nil)
		v := NewVolumeFromContext(tt.MountPoint, tt.Size, volumeID)
		if v.Name != tt.Name {
			t.Errorf("NewVolumeFromContext, volume's name: (%v), want (%v)", v.Name, tt.Name)
		}

		if v.Driver() != tt.Driver {
			t.Errorf("NewVolumeFromContext, volume's driver: (%v), want (%v)", v.Driver(), tt.Driver)
		}

		if v.Path() != tt.MountPoint {
			t.Errorf("NewVolumeFromContext, volume's driver: (%v), want (%v)", v.Path(), tt.MountPoint)
		}

		if v.Size() != tt.Size {
			t.Errorf("NewVolumeFromContext, volume's size: (%v), want (%v)", v.Size(), tt.Size)
		}
	}
}
