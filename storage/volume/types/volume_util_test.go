package types

import "testing"

func TestNewVolumeFromID(t *testing.T) {
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
		volumeID := NewVolumeID(tt.Name, tt.Driver)
		v := NewVolumeFromID(tt.MountPoint, tt.Size, volumeID)
		if v.Name != tt.Name {
			t.Errorf("NewVolumeFromID, volume's name: (%v), want (%v)", v.Name, tt.Name)
		}

		if v.Driver() != tt.Driver {
			t.Errorf("NewVolumeFromID, volume's driver: (%v), want (%v)", v.Driver(), tt.Driver)
		}

		if v.Path() != tt.MountPoint {
			t.Errorf("NewVolumeFromID, volume's driver: (%v), want (%v)", v.Path(), tt.MountPoint)
		}

		if v.Size() != tt.Size {
			t.Errorf("NewVolumeFromID, volume's size: (%v), want (%v)", v.Size(), tt.Size)
		}
	}
}
