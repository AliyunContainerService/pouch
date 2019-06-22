package mgr

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/sirupsen/logrus"
)

// Kill kills a container.
func (mgr *ContainerManager) Kill(ctx context.Context, name string, signal uint64) (err error) {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	if syscall.Signal(signal) == syscall.SIGKILL {
		return mgr.kill(ctx, c)
	}
	return mgr.killWithSignal(ctx, c, int(signal))
}

func (mgr *ContainerManager) kill(ctx context.Context, c *Container) error {
	if !c.IsRunning() {
		return fmt.Errorf("Container %s is not running", c.ID)
	}

	if err := mgr.killDeadProcess(ctx, c, int(syscall.SIGKILL)); err != nil {
		if errtypes.IsNoSuchProcess(err) {
			return nil
		}

		if c.IsRunning() {
			return err
		}
	}

	if pid := c.State.Pid; pid != 0 {
		if err := syscall.Kill(int(pid), 9); err != nil {
			if err != syscall.ESRCH {
				return err
			}

			e := errors.Wrapf(errtypes.ErrNoSuchProcess, "Cannot kill process (pid=%d) with signal %d", c.State.Pid, 9)
			logrus.Debug(e)
			return nil
		}
	}
	return nil
}

func (mgr *ContainerManager) killDeadProcess(ctx context.Context, c *Container, signal int) error {
	err := mgr.killWithSignal(ctx, c, signal)
	if err == syscall.ESRCH {
		e := errors.Wrapf(errtypes.ErrNoSuchProcess, "Cannot kill process (pid=%d) with signal %d", c.State.Pid, signal)
		logrus.Debug(e)
		return e
	}
	return err
}

func (mgr *ContainerManager) killWithSignal(ctx context.Context, c *Container, signal int) error {
	logrus.Debugf("Sending %d to %s", signal, c.ID)
	c.Lock()
	defer c.Unlock()

	if c.State.Paused {
		return fmt.Errorf("Container %s is paused. Unpause the container before stopping", c.ID)
	}

	if !c.State.Running {
		return fmt.Errorf("Container %s is not running", c.ID)
	}

	if err := mgr.Client.KillContainer(ctx, c.ID, signal); err != nil {
		// if container or process not exists, ignore the error
		if strings.Contains(err.Error(), "container not found") ||
			strings.Contains(err.Error(), "no such process") {
			logrus.Warnf("container kill failed because of 'container not found' or 'no such process': %s", err.Error())
		} else {
			return fmt.Errorf("Cannot kill container %s: %s", c.ID, err)
		}
	}
	attributes := map[string]string{
		"signal": fmt.Sprintf("%d", signal),
	}
	mgr.LogContainerEventWithAttributes(ctx, c, "kill", attributes)
	return nil
}
