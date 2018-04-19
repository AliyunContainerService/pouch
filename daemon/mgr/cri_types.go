package mgr

import (
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// SandboxMeta represents the sandbox's meta data.
type SandboxMeta struct {
	// ID is the id of sandbox.
	ID string

	// Config is CRI sandbox config.
	Config *runtime.PodSandboxConfig

	// NetNSPath is the network namespace used by the sandbox.
	NetNSPath string
}

// Key returns sandbox's id.
func (meta *SandboxMeta) Key() string {
	return meta.ID
}
