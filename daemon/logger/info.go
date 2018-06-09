package logger

import "github.com/alibaba/pouch/pkg/utils"

// Info provides container information for log driver.
type Info struct {
	LogConfig map[string]string

	ContainerID      string
	ContainerLabels  map[string]string
	ContainerRootDir string
}

// ID returns the container truncated ID.
func (i *Info) ID() string {
	return utils.TruncateID(i.ContainerID)
}
