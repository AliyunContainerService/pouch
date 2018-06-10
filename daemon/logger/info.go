package logger

import (
	"github.com/alibaba/pouch/pkg/utils"
)

// Info provides container information for log driver.
//
// TODO(fuwei): add more fields.
type Info struct {
	LogConfig map[string]string

	ContainerID      string
	ContainerName    string
	ContainerImageID string
	ContainerLabels  map[string]string
	ContainerRootDir string

	DaemonName string
}

// ID returns the container truncated ID.
func (i *Info) ID() string {
	return utils.TruncateID(i.ContainerID)
}

// FullID returns the container ID.
func (i *Info) FullID() string {
	return i.ContainerID
}

// Name returns the container name.
func (i *Info) Name() string {
	return i.ContainerName
}

// ImageID returns the container's image truncated ID.
func (i *Info) ImageID() string {
	return utils.TruncateID(i.ContainerImageID)
}

// ImageFullID returns the container's image ID.
func (i *Info) ImageFullID() string {
	return i.ContainerImageID
}
