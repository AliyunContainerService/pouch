package mgr

import (
	"os"
	"path/filepath"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/docker/docker/pkg/symlink"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

// IsRemoving returns container is removing or not.
func (c *Container) IsRemoving() bool {
	c.Lock()
	defer c.Unlock()
	return c.State.Status == types.StatusRemoving
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
	c.setStatusFlags(types.StatusRunning)
}

// StatPath is the unexported version of StatPath. Locks and mounts should
// be acquired before calling this method and the given path should be fully
// resolved to a path on the host corresponding to the given absolute path
// inside the container.
func (c *Container) StatPath(resolvedPath, absPath string) (stat *types.ContainerPathStat, err error) {
	lstat, err := os.Lstat(resolvedPath)
	if err != nil {
		return nil, errors.Wrap(errtypes.ErrNotfound, "container "+c.Name)
	}

	var linkTarget string
	if lstat.Mode()&os.ModeSymlink != 0 {
		// Fully evaluate the symlink in the scope of the container rootfs.
		hostPath, err := c.GetResourcePath(absPath)
		if err != nil {
			return nil, errors.Wrap(errtypes.ErrNotfound, "container "+c.Name)
		}

		linkTarget, err = filepath.Rel(c.BaseFS, hostPath)
		if err != nil {

			return nil, errors.Wrap(errtypes.ErrNotfound, "container "+c.Name)
		}

		// Make it an absolute path.
		linkTarget = filepath.Join(string(filepath.Separator), linkTarget)
	}

	return &types.ContainerPathStat{
		Name:       filepath.Base(absPath),
		Size:       lstat.Size(),
		Mode:       uint32(lstat.Mode()),
		Mtime:      strfmt.DateTime(lstat.ModTime()),
		LinkTarget: linkTarget,
	}, nil
}

// GetResourcePath evaluates `path` in the scope of the container's BaseFS, with proper path
// sanitisation. Symlinks are all scoped to the BaseFS of the container, as
// though the container's BaseFS was `/`.
//
// The BaseFS of a container is the host-facing path which is bind-mounted as
// `/` inside the container. This method is essentially used to access a
// particular path inside the container as though you were a process in that
// container.
//
// NOTE: The returned path is *only* safely scoped inside the container's BaseFS
//       if no component of the returned path changes (such as a component
//       symlinking to a different path) between using this method and using the
//       path. See symlink.FollowSymlinkInScope for more details.
func (c *Container) GetResourcePath(path string) (string, error) {
	// IMPORTANT - These are paths on the OS where the daemon is running, hence
	// any filepath operations must be done in an OS agnostic way.
	cleanPath := cleanResourcePath(path)
	r, e := symlink.FollowSymlinkInScope(filepath.Join(c.BaseFS, cleanPath), c.BaseFS)

	// Log this here on the daemon side as there's otherwise no indication apart
	// from the error being propagated all the way back to the client. This makes
	// debugging significantly easier and clearly indicates the error comes from the daemon.
	if e != nil {
		logrus.Errorf("Failed to FollowSymlinkInScope BaseFS %s cleanPath %s path %s %s\n", c.BaseFS, cleanPath, path, e)
	}
	return r, e
}

// cleanResourcePath cleans a resource path and prepares to combine with mnt path
func cleanResourcePath(path string) string {
	return filepath.Join(string(os.PathSeparator), path)
}

// SetStatusStopped sets a container to be status stopped.
// When a container's status turns to StatusStopped, the following fields need updated:
// Status -> StatusStopped
// FinishedAt -> time.Now()
// Pid -> 0
// ExitCode -> input param
// Error -> input param
func (c *Container) SetStatusStopped(exitCode int64, errMsg string) {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusStopped
	c.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = 0
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.setStatusFlags(types.StatusStopped)
}

// SetStatusExited sets a container to be status exited.
func (c *Container) SetStatusExited(exitCode int64, errMsg string) {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusExited
	c.State.FinishedAt = time.Now().UTC().Format(utils.TimeLayout)
	c.State.Pid = 0
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.setStatusFlags(types.StatusExited)
}

// SetStatusPaused sets a container to be status paused.
func (c *Container) SetStatusPaused() {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusPaused
	c.setStatusFlags(types.StatusPaused)
}

// SetStatusUnpaused sets a container to be status running.
// Unpaused is treated running.
func (c *Container) SetStatusUnpaused() {
	c.Lock()
	defer c.Unlock()
	c.State.Status = types.StatusRunning
	c.setStatusFlags(types.StatusRunning)
}

// Notes(ziren): i still feel uncomfortable for a function hasing no return
// setStatusFlags set the specified status flag to true, and unset others
func (c *Container) setStatusFlags(status types.Status) {
	statusFlags := map[types.Status]bool{
		types.StatusDead:       false,
		types.StatusRunning:    false,
		types.StatusPaused:     false,
		types.StatusRestarting: false,
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
		}
	}
}
