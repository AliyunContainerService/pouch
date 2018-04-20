package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// SpecWrapper wraps the container's specs and add manager operations.
type SpecWrapper struct {
	s *specs.Spec

	ctrMgr  ContainerMgr
	volMgr  VolumeMgr
	netMgr  NetworkMgr
	prioArr []int
	argsArr [][]string
}

// SetupFunc defines spec setup function type.
type SetupFunc func(ctx context.Context, m *ContainerMeta, s *SpecWrapper) error

var setupFunc = []SetupFunc{
	// process
	setupProcessArgs,
	setupProcessCwd,
	setupProcessEnv,
	setupProcessTTY,
	setupProcessUser,
	setupCap,
	setupNoNewPrivileges,
	setupOOMScoreAdj,

	// cgroup
	setupCgroupCPUShare,
	setupCgroupCPUSet,
	setupCgroupCPUPeriod,
	setupCgroupCPUQuota,
	setupCgroupMemory,
	setupCgroupMemorySwap,
	setupCgroupMemorySwappiness,
	setupDisableOOMKill,

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

	// host device spec
	setupDevices,

	// linux-platform-specifc spec
	setupSysctl,
	setupAppArmor,
	setupCapabilities,
	setupSeccomp,
	setupSELinux,

	// blkio spec
	setupBlkio,
	setupDiskQuota,

	// IntelRdtL3Cbm
	setupIntelRdt,

	// annotations in spec
	setupAnnotations,

	// rootfs spec
	setupRoot,

	//hook
	setupHook,
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
