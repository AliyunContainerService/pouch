package annotations

// ContainerType values
const (
	// ContainerTypeSandbox represents a pod sandbox container
	ContainerTypeSandbox = "sandbox"

	// ContainerTypeContainer represents a container running within a pod
	ContainerTypeContainer = "container"

	// CRIOContainerType is the container type (sandbox or container) annotation
	CRIOContainerType = "io.kubernetes.cri-o.ContainerType"

	// ContainerType is the container type (sandbox or container) annotation
	ContainerType = "io.kubernetes.cri.container-type"

	// CRIOSandboxName is the sandbox name annotation
	CRIOSandboxName = "io.kubernetes.cri-o.SandboxName"

	// CRIOSandboxID is the sandbox id annotation
	CRIOSandboxID = "io.kubernetes.cri-o.SandboxID"

	// SandboxID is the sandbox ID annotation
	SandboxID = "io.kubernetes.cri.sandbox-id"

	// KubernetesRuntime is the runtime
	KubernetesRuntime = "io.kubernetes.runtime"

	// LxcfsEnabled whether to enable lxcfs for a container
	LxcfsEnabled = "io.kubernetes.lxcfs.enabled"

	// ExtendAnnotationPrefix is the extend annotation prefix
	ExtendAnnotationPrefix = "io.alibaba.pouch"

	// MemorySwapExtendAnnotation is the extend annotation of memory swap
	MemorySwapExtendAnnotation = "io.alibaba.pouch.resources.memory-swap"

	// PidsLimitExtendAnnotation is the extend annotation of pids limit
	PidsLimitExtendAnnotation = "io.alibaba.pouch.resources.pids-limit"
)
