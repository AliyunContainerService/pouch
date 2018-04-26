package volume

import (
	"time"
)

// Config represents volume config struct.
type Config struct {
	ControlAddress string
	Timeout        time.Duration // operation timeout.
	RemoveVolume   bool
	DefaultBackend string `json:"volume-default-driver"`
	VolumeMetaPath string `json:"volume-meta-dir"`
	DriverAlias    string `json:"volume-driver-alias"`
}
