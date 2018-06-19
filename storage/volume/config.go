package volume

import (
	"time"
)

// Config represents volume config struct.
type Config struct {
	Timeout        time.Duration `json:"volume-timeout,omitempty"`        // operation timeout.
	RemoveVolume   bool          `json:"remove-volume,omitempty"`         // remove volume add data or volume's metadata when remove pouch volume.
	DefaultBackend string        `json:"volume-default-driver,omitempty"` // default volume backend.
	VolumeMetaPath string        `json:"volume-meta-dir,omitempty"`       // volume metadata store path.
	DriverAlias    string        `json:"volume-driver-alias,omitempty"`   // driver alias configure.
}
