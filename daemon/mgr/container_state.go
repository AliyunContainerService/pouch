package mgr

import (
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
)

// IsRunning returns container is running or not.
func (c *Container) IsRunning() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusRunning
}

// IsStopped returns container is stopped or not.
func (c *Container) IsStopped() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusStopped
}

// IsExited returns container is exited or not.
func (c *Container) IsExited() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusExited
}

// IsCreated returns container is created or not.
func (c *Container) IsCreated() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusCreated
}

// IsPaused returns container is paused or not.
func (c *Container) IsPaused() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusPaused
}

// IsDead returns container is dead or not.
func (c *Container) IsDead() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusDead
}

// IsRunningOrPaused returns true of container is running or paused.
func (c *Container) IsRunningOrPaused() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusRunning || c.State.Status == types.StatusPaused
}

// IsRestarting returns container is restarting or not.
func (c *Container) IsRestarting() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusRestarting
}

// ExitCode returns container's ExitCode.
func (c *Container) ExitCode() int64 {
	c.Lock()
	defer c.Unlock()
	return c.State.ExitCode
}

// SetStatusRunning sets a container to be status running.
// When a container's status turns to StatusStopped, the following fields need updated:
// Status -> StatusRunning
// StartAt -> time.Now()
// Pid -> input param
// ExitCode -> 0
func (c *Container) SetStatusRunning(pid int64) {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusRunning
	c.State.StartedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = pid
	c.State.ExitCode = 0
}

// SetStatusStopped sets a container to be status stopped.
// When a container's status turns to StatusStopped, the following fields need updated:
// Status -> StatusStopped
// FinishedAt -> time.Now()
// Pid -> -1
// ExitCode -> input param
// Error -> input param
func (c *Container) SetStatusStopped(exitCode int64, errMsg string) {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusStopped
	c.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = -1
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
}

// SetStatusExited sets a container to be status exited.
func (c *Container) SetStatusExited() {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusExited
}

// SetStatusPaused sets a container to be status paused.
func (c *Container) SetStatusPaused() {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusPaused
}

// SetStatusUnpaused sets a container to be status running.
// Unpaused is treated running.
func (c *Container) SetStatusUnpaused() {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusRunning
}
