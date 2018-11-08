package v1alpha2

import (
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
)

// SandboxMeta represents the sandbox's meta data.
type SandboxMeta struct {
	// ID is the id of sandbox.
	ID string

	// SandboxContainerID is the container id of sandbox.
	SandboxContainerID string

	// Config is CRI sandbox config.
	Config *runtime.PodSandboxConfig

	// Runtime is the runtime of sandbox
	Runtime string

	// Runtime whether to enable lxcfs for a container
	LxcfsEnabled bool

	// NetNS is the sandbox's network namespace
	NetNS string
}

// Key returns sandbox's id.
func (meta *SandboxMeta) Key() string {
	return meta.ID
}
