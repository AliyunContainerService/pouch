package mgr

import (
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func Test_setNvidiaHook(t *testing.T) {
	nvidiaHookName = "test-nvidia-container-runtime-hook"
	installDir := "/usr/local/bin"
	fullname := path.Join(installDir, nvidiaHookName)
	nvidiaHookName = fullname
	os.Remove(fullname)
	os.Create(fullname)
	os.Chmod(fullname, 0755)
	hookPath := ""
	if !path.IsAbs(nvidiaHookName) {
		hookPath, _ = exec.LookPath(nvidiaHookName)
	} else {
		hookPath = nvidiaHookName
	}
	defer func() {
		os.Remove(fullname)
	}()

	tests := []struct {
		name             string
		c                *Container
		specWrapper      *SpecWrapper
		expectedPrestart []specs.Hook
	}{
		{
			"NvidiaConfig is nil, NvidiaEnv is null",
			&Container{
				HostConfig: &types.HostConfig{
					Resources: types.Resources{
						NvidiaConfig: nil,
					},
				},
				Config: &types.ContainerConfig{
					Env: []string{},
				},
			},
			&SpecWrapper{
				s: &specs.Spec{
					Hooks: &specs.Hooks{
						Prestart: []specs.Hook{},
					},
				},
			},
			[]specs.Hook{},
		},
		{
			"NvidiaConfig is nil, NvidiaEnv not null",
			&Container{
				HostConfig: &types.HostConfig{
					Resources: types.Resources{
						NvidiaConfig: nil,
					},
				},
				Config: &types.ContainerConfig{
					Env: []string{"NVIDIA_DRIVER_CAPABILITIES=all", "NVIDIA_VISIBLE_DEVICES=all"},
				},
			},
			&SpecWrapper{
				s: &specs.Spec{
					Hooks: &specs.Hooks{
						Prestart: []specs.Hook{},
					},
				},
			},
			[]specs.Hook{{
				Path: hookPath,
				Args: append([]string{hookPath}, "prestart"),
			}},
		},
		{
			"NvidiaConfig not nil, NvidiaEnv is null",
			&Container{
				HostConfig: &types.HostConfig{
					Resources: types.Resources{
						NvidiaConfig: &types.NvidiaConfig{
							NvidiaDriverCapabilities: "all",
							NvidiaVisibleDevices:     "all",
						},
					},
				},
				Config: &types.ContainerConfig{
					Env: []string{},
				},
			},
			&SpecWrapper{
				s: &specs.Spec{
					Hooks: &specs.Hooks{
						Prestart: []specs.Hook{},
					},
				},
			},
			[]specs.Hook{{
				Path: hookPath,
				Args: append([]string{hookPath}, "prestart"),
			}},
		},
	}
	for _, tt := range tests {
		err := setNvidiaHook(tt.c, tt.specWrapper)
		if err != nil {
			t.Errorf("setNvidiaHook = %v, want %v", err, nil)
		}
		if !reflect.DeepEqual(tt.specWrapper.s.Hooks.Prestart, tt.expectedPrestart) {
			t.Errorf("setNvidiaHook = %v, want %v", tt.specWrapper.s.Hooks.Poststart, tt.expectedPrestart)
		}
	}
}
