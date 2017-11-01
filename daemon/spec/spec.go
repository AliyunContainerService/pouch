package spec

import (
	"context"

	"github.com/alibaba/pouch/apis/types"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// SetupFunc defines spec setup function type.
type SetupFunc func(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error

var setupFunc []SetupFunc

// RegisteSetupFunc is used to registe spec setup function.
func RegisteSetupFunc(f SetupFunc) {
	if setupFunc == nil {
		setupFunc = make([]SetupFunc, 0)
	}
	setupFunc = append(setupFunc, f)
}

// GetSetupFunc returns all the spec setup functions.
func GetSetupFunc() []SetupFunc {
	return setupFunc
}
