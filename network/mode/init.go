package mode

import (
	"context"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/network/mode/bridge"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NetworkModeInit is used to initilize network mode, include host and none network.
func NetworkModeInit(ctx context.Context, config network.Config, manager mgr.NetworkMgr) error {
	// if it has old containers, don't to intialize network.
	if len(config.ActiveSandboxes) > 0 {
		logrus.Warnf("There are old containers, don't to initialize network")
		return nil
	}

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
	if !config.BridgeConfig.DisableBridge {
		if err := bridge.New(ctx, config.BridgeConfig, manager); err != nil {
			return errors.Wrapf(err, "failed to init bridge network")
		}
	}
	return nil
}
