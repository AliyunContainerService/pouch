package mgr

import (
	"context"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

//if set rich mode, set initscript
func setupHook(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	if !c.Config.Rich || c.Config.InitScript == "" {
		return nil
	}

	args := strings.Fields(c.Config.InitScript)
	if len(args) == 0 {
		return nil
	}

	if spec.s.Hooks == nil {
		spec.s.Hooks = &specs.Hooks{}
	}

	if spec.s.Hooks.Prestart == nil {
		spec.s.Hooks.Prestart = []specs.Hook{}
	}

	preStartHook := specs.Hook{
		Path: args[0],
		Args: args[1:],
	}

	spec.s.Hooks.Prestart = append(spec.s.Hooks.Prestart, preStartHook)

	return nil
}
