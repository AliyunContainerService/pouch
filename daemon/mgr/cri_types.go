package mgr

import (
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// SandboxMeta represents the sandbox's meta data.
type SandboxMeta struct {
	// Config is CRI sandbox config.
	Config *runtime.PodSandboxConfig

	// NetNSPath is the network namespace used by the sandbox.
	NetNSPath string
}
