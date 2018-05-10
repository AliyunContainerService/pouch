package volume

import (
	"time"
)

// Config represents volume config struct.
type Config struct {
	ControlServerAddress string        `json:"control-server-address"` // control server address in csi.
	Timeout              time.Duration `json:"volume-timeout"`         // operation timeout.
	RemoveVolume         bool          `json:"remove-volume"`          // remove volume add data or volume's metadata when remove pouch volume.
	DefaultBackend       string        `json:"volume-default-driver"`  // default volume backend.
	VolumeMetaPath       string        `json:"volume-meta-dir"`        // volume metadata store path.
	DriverAlias          string        `json:"volume-driver-alias"`    // driver alias configure.
}
