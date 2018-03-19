package mgr

import (
	"context"
	"sort"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

//setup hooks specified by user via plugins, if set rich mode and init-script exists set init-script
func setupHook(ctx context.Context, c *ContainerMeta, spec *SpecWrapper) error {
	if len(spec.argsArr) > 0 {
		var hookArr []*wrapperEmbedPrestart
		for i, hook := range spec.s.Hooks.Prestart {
			hookArr = append(hookArr, &wrapperEmbedPrestart{-i, append([]string{hook.Path}, hook.Args...)})
		}
		priorityArr := spec.prioArr
		argsArr := spec.argsArr
		for i, p := range priorityArr {
			hookArr = append(hookArr, &wrapperEmbedPrestart{p, argsArr[i]})
		}
		sortedArr := hookArray(hookArr)
		sort.Sort(sortedArr)
		spec.s.Hooks.Prestart = sortedArr.toOciPrestartHook()
	}

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

type hookArray []*wrapperEmbedPrestart

// Len is defined in order to support sort
func (h hookArray) Len() int {
	return len(h)
}

// Len is defined in order to support sort
func (h hookArray) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Less is defined in order to support sort, bigger priority execute first
func (h hookArray) Less(i, j int) bool {
	return h[i].Priority()-h[j].Priority() > 0
}

func (h hookArray) toOciPrestartHook() []specs.Hook {
	allHooks := make([]specs.Hook, len(h))
	for i, hook := range h {
		allHooks[i].Path = hook.Hook()[0]
		allHooks[i].Args = hook.Hook()[1:]
	}
	return allHooks
}

type wrapperEmbedPrestart struct {
	p    int
	args []string
}

func (w *wrapperEmbedPrestart) Priority() int {
	return w.p
}

func (w *wrapperEmbedPrestart) Hook() []string {
	return w.args
}
