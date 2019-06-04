package mgr

import (
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
)

// IsRunning returns container is running or not.
func (c *Container) IsRunning() bool {
	return c.State.Running
}

// IsRunningOrPaused returns true of container is running or paused.
func (c *Container) IsRunningOrPaused() bool {
	return c.State.Running || c.State.Paused
}

// ExitCode returns container's ExitCode.
func (c *Container) ExitCode() int64 {
	return c.State.ExitCode
}

// IsCreated returns container is created or not.
func (c *Container) IsCreated() bool {
	return c.State.Status == types.StatusCreated
}

// IsRemoving returns container is removing or not.
// TODO: actually the pouchd do not set removing status for a container.
func (c *Container) IsRemoving() bool {
	return c.State.Status == types.StatusRemoving
}

// IsStopped returns container is stopped or not.
func (c *Container) IsStopped() bool {
	return c.State.Status == types.StatusStopped
}

// SetStatusRunning sets a container to be status running.
// When a container's status turns to StatusStopped, the following fields need updated:
// Status -> StatusRunning
// StartAt -> time.Now()
// Pid -> input param
// ExitCode -> 0
func (c *Container) SetStatusRunning(pid int64) {
	c.State.Status = types.StatusRunning
	c.State.StartedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = pid
	c.State.ExitCode = 0
	c.setStatusFlags(types.StatusRunning)
}

// SetStatusStopped sets a container to be status stopped.
// When a container's status turns to StatusStopped, the following fields need updated:
// Status -> StatusStopped
// FinishedAt -> time.Now()
// Pid -> 0
// ExitCode -> input param
// Error -> input param
func (c *Container) SetStatusStopped(exitCode int64, errMsg string) {
	c.State.Status = types.StatusStopped
	c.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = 0
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.setStatusFlags(types.StatusStopped)
}

// SetStatusExited sets a container to be status exited.
func (c *Container) SetStatusExited(exitCode int64, errMsg string) {
	c.State.Status = types.StatusExited
	c.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = 0
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.setStatusFlags(types.StatusExited)
}

// SetStatusPaused sets a container to be status paused.
func (c *Container) SetStatusPaused() {
	c.State.Status = types.StatusPaused
	c.setStatusFlags(types.StatusPaused)
}

// SetStatusUnpaused sets a container to be status running.
// Unpaused is treated running.
func (c *Container) SetStatusUnpaused() {
	c.State.Status = types.StatusRunning
	c.setStatusFlags(types.StatusRunning)
}

// SetStatusOOM sets a container to be status exit because of OOM.
func (c *Container) SetStatusOOM() {
	c.State.OOMKilled = true
	c.State.Error = "OOMKilled"
}

// Notes(ziren): i still feel uncomfortable for a function hasing no return
// setStatusFlags set the specified status flag to true, and unset others
func (c *Container) setStatusFlags(status types.Status) {
	statusFlags := map[types.Status]bool{
		types.StatusDead:       false,
		types.StatusRunning:    false,
		types.StatusPaused:     false,
		types.StatusRestarting: false,
		types.StatusExited:     false,
	}

	if _, exists := statusFlags[status]; exists {
		statusFlags[status] = true
	}

	for k, v := range statusFlags {
		switch k {
		case types.StatusDead:
			c.State.Dead = v
		case types.StatusPaused:
			c.State.Paused = v
		case types.StatusRunning:
			c.State.Running = v
		case types.StatusRestarting:
			c.State.Restarting = v
		case types.StatusExited:
			c.State.Exited = v
		}
	}
}
