package mgr

import (
	"context"
	"os/exec"
	"sort"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

//setup hooks specified by user via plugins, if set rich mode and init-script exists set init-script
func setupHook(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	if s.Hooks == nil {
		s.Hooks = &specs.Hooks{
			Prestart:  []specs.Hook{},
			Poststart: []specs.Hook{},
			Poststop:  []specs.Hook{},
		}
	}

	// setup plugin hook, if no hook plugin setup, skip this part.
	argsArr := specWrapper.argsArr
	prioArr := specWrapper.prioArr
	if len(argsArr) > 0 {
		var hookArr []*wrapperEmbedPrestart
		for i, hook := range s.Hooks.Prestart {
			hookArr = append(hookArr, &wrapperEmbedPrestart{-i, append([]string{hook.Path}, hook.Args...)})
		}
		for i, p := range prioArr {
			hookArr = append(hookArr, &wrapperEmbedPrestart{p, argsArr[i]})
		}
		sortedArr := hookArray(hookArr)
		sort.Sort(sortedArr)
		s.Hooks.Prestart = append(s.Hooks.Prestart, sortedArr.toOciPrestartHook()...)
	}

	// setup rich mode container hoopk, if no init script specified and no hook plugin setup, skip this part.
	if c.Config.Rich && c.Config.InitScript != "" {
		args := strings.Fields(c.Config.InitScript)
		if len(args) > 0 {
			s.Hooks.Prestart = append(s.Hooks.Prestart, specs.Hook{
				Path: args[0],
				Args: args[1:],
			})
		}
	}

	// set nvidia config
	if err := setNvidiaHook(ctx, c, specWrapper); err != nil {
		return errors.Wrap(err, "failed to set nvidia prestart hook")
	}

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

func setNvidiaHook(ctx context.Context, c *Container, spec *SpecWrapper) error {
	n := c.HostConfig.NvidiaConfig
	if n == nil {
		return nil
	}
	path, err := exec.LookPath("nvidia-container-runtime-hook")
	if err != nil {
		return err
	}
	args := []string{path}
	nvidiaPrestart := specs.Hook{
		Path: path,
		Args: append(args, "prestart"),
	}
	spec.s.Hooks.Prestart = append(spec.s.Hooks.Prestart, nvidiaPrestart)
	return nil
}
