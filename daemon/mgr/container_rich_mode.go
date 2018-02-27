package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
)

const (
	richModeEnv       = "rich_mode=true"
	richModeLaunchEnv = "rich_mode_launch_manner"

	//for ali internal
	interRichModeEnv = "ali_run_mode=common_vm"
)

func richContainerModeEnv(c *ContainerMeta) []string {
	var (
		ret         = []string{}
		setRichMode = false
		richMode    = ""
	)

	envs := c.Config.Env

	//if set inter_rich_mode_env, you can also run in rich container mode
	for _, e := range envs {
		if e == interRichModeEnv {
			setRichMode = true
			richMode = types.ContainerConfigRichModeSystemd
			break
		}
	}

	if c.Config.Rich {
		setRichMode = true
	}

	if c.Config.RichMode != "" {
		richMode = c.Config.RichMode
	}

	//if not set rich mode manner, default set dumb-init
	if richMode == "" {
		richMode = types.ContainerConfigRichModeDumbInit
	}

	if setRichMode {
		ret = append(ret, richModeEnv, fmt.Sprintf("%s=%s", richModeLaunchEnv, richMode))
	}

	return ret
}
