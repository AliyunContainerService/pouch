package mgr

import (
	"os/exec"
	"path"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	// nvidiaHookName is a custom OCI prestart hook binary to runc in order to enable GPU containers.
	nvidiaHookName = "nvidia-container-runtime-hook"
)

func setNvidiaHook(c *Container, spec *SpecWrapper) error {
	n := c.HostConfig.NvidiaConfig

	// to make compatible for k8s.
	// if user set environments of NVIDIA, then set prestart hook
	kv := utils.ConvertKVStrToMapWithNoErr(c.Config.Env)
	_, hasEnvCapabilities := kv["NVIDIA_DRIVER_CAPABILITIES"]
	_, hasEnvDevices := kv["NVIDIA_VISIBLE_DEVICES"]

	if n == nil && !hasEnvCapabilities && !hasEnvDevices {
		return nil
	}

	hookPath := ""
	if !path.IsAbs(nvidiaHookName) {
		var err error
		hookPath, err = exec.LookPath(nvidiaHookName)
		if err != nil {
			return err
		}
	} else {
		hookPath = nvidiaHookName
	}
	args := []string{hookPath}
	nvidiaPrestart := specs.Hook{
		Path: hookPath,
		Args: append(args, "prestart"),
	}
	spec.s.Hooks.Prestart = append(spec.s.Hooks.Prestart, nvidiaPrestart)

	return nil
}
