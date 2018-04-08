package mgr

import (
	"github.com/alibaba/pouch/apis/types"
)

// verifyContainerSetting is to verify the correctness of hostconfig and config.
func (mgr *ContainerManager) verifyContainerSetting(hostConfig *types.HostConfig, config *types.ContainerConfig) error {
	if config != nil {
		// TODO
	}
	return nil
}
