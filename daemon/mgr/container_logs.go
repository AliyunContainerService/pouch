package mgr

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/jsonfile"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/utils"

	pkgerrors "github.com/pkg/errors"
)

var watchTimeout = 300 * time.Millisecond

// Logs is used to return log created by the container.
func (mgr *ContainerManager) Logs(ctx context.Context, name string, logOpt *types.ContainerLogsOptions) (<-chan *logger.LogMessage, bool, error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, false, err
	}

	if !(logOpt.ShowStdout || logOpt.ShowStderr) {
		return nil, false, pkgerrors.Wrap(errtypes.ErrInvalidParam, "you must choose at least one stream")
	}

	if c.HostConfig.LogConfig.LogDriver != types.LogConfigLogDriverJSONFile {
		return nil, false, pkgerrors.Wrapf(
			errtypes.ErrInvalidParam,
			"only support for the %v log driver", types.LogConfigLogDriverJSONFile,
		)
	}

	cfg, err := convContainerLogsOptionsToReadConfig(logOpt)
	if err != nil {
		return nil, false, err
	}

	// NOTE: created container doesn't create IO.
	if c.IsCreated() {
		msgCh := make(chan *logger.LogMessage, 1)
		close(msgCh)

		return msgCh, c.Config.Tty, nil
	}

	fileName := filepath.Join(mgr.Store.Path(c.ID), "json.log")
	jf, err := jsonfile.NewJSONLogFile(fileName, 0640)
	if err != nil {
		return nil, false, err
	}

	// NOTE: unset the follow if the container is not running
	cfg.Follow = cfg.Follow && c.IsRunning()

	msgCh := make(chan *logger.LogMessage, 1)
	watcher := jf.ReadLogMessages(cfg)

	go func() {
		defer jf.Close()
		defer watcher.Close()
		defer close(msgCh)

		// FIXME: in current design, we cannot reuse the existing
		// containerio to create/notify all the related watcher.
		// for the follow case, if the container has been stopped, we
		// should return. There is only way to use timer to spin checking
		// the status of container.
		watchTimer := time.NewTimer(time.Second)
		defer watchTimer.Stop()

		for {
			watchTimer.Reset(watchTimeout)
			select {
			case <-ctx.Done():
				return
			case err := <-watcher.Err:
				select {
				case <-ctx.Done():
					return
				case msgCh <- &logger.LogMessage{Err: err}:
					return
				}
			case msg, ok := <-watcher.Msgs:
				if !ok {
					// NOTE: channel closed by the ReadLogMessages
					return
				}

				select {
				case <-ctx.Done():
					return
				case msgCh <- msg:
				}
			case <-watchTimer.C:
				// NOTE: if it is not OK, it maybe removed.
				// This case will be convered by the followFile
				// in daemon/logger/jsonfile package.
				if c, ok := mgr.cache.Get(c.ID).Result(); ok {
					if !c.(*Container).IsRunning() {
						return
					}
				}

			}
		}
	}()
	return msgCh, c.Config.Tty, nil
}

func convContainerLogsOptionsToReadConfig(logOpt *types.ContainerLogsOptions) (*logger.ReadConfig, error) {
	var since time.Time
	if logOpt.Since != "" {
		sec, nano, err := utils.ParseTimestamp(logOpt.Since, 0)
		if err != nil {
			return nil, err
		}
		since = time.Unix(sec, nano)
	}

	var until time.Time
	if logOpt.Until != "" {
		sec, nano, err := utils.ParseTimestamp(logOpt.Until, 0)
		if err != nil {
			return nil, err
		}
		until = time.Unix(sec, nano)
	}

	lines, err := strconv.Atoi(logOpt.Tail)
	if err != nil {
		lines = -1
	}

	return &logger.ReadConfig{
		Since:  since,
		Until:  until,
		Follow: logOpt.Follow,
		Tail:   lines,
	}, nil
}
