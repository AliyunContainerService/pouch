package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
)

const (
	richModeEnv       = "rich_mode=true"
	richModeLaunchEnv = "rich_mode_launch_manner"
)

func richContainerModeEnv(c *Container) []string {
	if !c.Config.Rich {
		return nil
	}

	if c.Config.RichMode == "" {
		c.Config.RichMode = types.ContainerConfigRichModeDumbInit
	}

	return []string{richModeEnv, fmt.Sprintf("%s=%s", richModeLaunchEnv, c.Config.RichMode)}
}
