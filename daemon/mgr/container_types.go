package mgr

import (
	"github.com/alibaba/pouch/apis/types"
)

type containerExecConfig struct {
	types.ExecCreateConfig

	// Save the container's id into exec config.
	ContainerID string
}
