package types

import (
	"time"
)

// Config represents volume config struct.
type Config struct {
	ControlAddress string
	Timeout        time.Duration // operation timeout.
	RemoveVolume   bool
	DefaultBackend string
	VolumeMetaPath string
}
