package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// SetupFunc defines spec setup function type.
type SetupFunc func(ctx context.Context, m *ContainerMeta, s *specs.Spec) error

var setupFunc = []SetupFunc{
	// process
	setupProcessArgs,
	setupProcessCwd,
	setupProcessEnv,
	setupProcessTTY,
	setupProcessUser,
	setupCap,

	// cgroup
	setupCgroupCPUShare,
	setupCgroupCPUSet,
	setupCgroupMemory,
	setupCgroupMemorySwap,
	setupCgroupMemorySwappiness,

	// namespaces
	setupUserNamespace,
	setupNetworkNamespace,
	setupIpcNamespace,
	setupPidNamespace,
	setupUtsNamespace,

	// volume spec
	setupMounts,

	// network spec
	setupNetwork,
}

// Register is used to registe spec setup function.
func Register(f SetupFunc) {
	if setupFunc == nil {
		setupFunc = make([]SetupFunc, 0)
	}
	setupFunc = append(setupFunc, f)
}

// SetupFuncs returns all the spec setup functions.
func SetupFuncs() []SetupFunc {
	return setupFunc
}
