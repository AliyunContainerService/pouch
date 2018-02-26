package mode

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/network/mode/bridge"
	"github.com/sirupsen/logrus"
)

// NetworkModeInit is used to initilize network mode, include host and none network.
func NetworkModeInit(ctx context.Context, config network.Config, manager mgr.NetworkMgr) error {
	// init none network
	if n, _ := manager.Get(ctx, "none"); n == nil {
		logrus.Debugf("create none network")

		networkCreate := types.NetworkCreate{
			Driver: "null",
			Options: map[string]string{
				"persist": "true",
			},
		}
		create := types.NetworkCreateConfig{
			Name:          "none",
			NetworkCreate: networkCreate,
		}
		if _, err := manager.Create(ctx, create); err != nil {
			logrus.Errorf("failed to create none network, err: %v", err)
			return err
		}
	}

	// init host network
	if n, _ := manager.Get(ctx, "host"); n == nil {
		logrus.Debugf("create host network")

		networkCreate := types.NetworkCreate{
			Driver: "host",
			Options: map[string]string{
				"persist": "true",
			},
		}
		create := types.NetworkCreateConfig{
			Name:          "host",
			NetworkCreate: networkCreate,
		}
		if _, err := manager.Create(ctx, create); err != nil {
			logrus.Errorf("failed to create host network, err: %v", err)
			return err
		}
	}

	// init bridge network
	return bridge.New(ctx, config.BridgeConfig, manager)
}
