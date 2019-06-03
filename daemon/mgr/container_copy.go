package mgr

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/ioutils"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/mount"
	"github.com/go-openapi/strfmt"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// StatPath stats the dir info at the specified path in the container.
func (mgr *ContainerManager) StatPath(ctx context.Context, name, path string) (stat *types.ContainerPathStat, err error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, err
	}
	c.Lock()
	defer c.Unlock()

	if c.State.Dead {
		return nil, pkgerrors.Errorf("container(%s) has been deleted", c.ID)
	}

	running := c.IsRunningOrPaused()
	err = mgr.Mount(ctx, c, false)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to mount cid(%s)", c.ID)
	}

	defer mgr.Unmount(ctx, c, false, !running)

	if !running {
		err = mgr.attachVolumes(ctx, c)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "failed to attachVolumes cid(%s)", c.ID)
		}
		defer mgr.detachVolumes(ctx, c, false)
	}

	err = c.mountVolumes()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to mountVolumes cid(%s)", c.ID)
	}
	defer c.unmountVolumes()

	resolvedPath, absPath := c.getResolvedPath(path)
	lstat, err := os.Lstat(resolvedPath)

	if err != nil {
		return nil, err
	}

	return &types.ContainerPathStat{
		Name:  lstat.Name(),
		Path:  absPath,
		Size:  strconv.FormatInt(lstat.Size(), 10),
		Mode:  uint32(lstat.Mode()),
		Mtime: strfmt.DateTime(lstat.ModTime()),
	}, nil
}

// ArchivePath return an archive and dir info at the specified path in the container.
func (mgr *ContainerManager) ArchivePath(ctx context.Context, name, path string) (content io.ReadCloser, stat *types.ContainerPathStat, err0 error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, nil, err
	}
	c.Lock()
	defer func() {
		if err0 != nil {
			c.Unlock()
		}
	}()

	if c.State.Dead {
		return nil, nil, pkgerrors.Errorf("container(%s) has been deleted", c.ID)
	}

	running := c.IsRunningOrPaused()
	err = mgr.Mount(ctx, c, false)
	if err != nil {
		return nil, nil, pkgerrors.Wrapf(err, "failed to mount cid(%s)", c.ID)
	}
	defer func() {
		if err0 != nil {
			mgr.Unmount(ctx, c, false, !running)
		}
	}()

	if !running {
		err = mgr.attachVolumes(ctx, c)
		if err != nil {
			return nil, nil, pkgerrors.Wrapf(err, "failed to attachVolumes cid(%s)", c.ID)
		}
		defer func() {
			if err0 != nil {
				mgr.detachVolumes(ctx, c, false)
			}
		}()
	}

	err = c.mountVolumes()
	if err != nil {
		return nil, nil, pkgerrors.Wrapf(err, "failed to mountVolumes cid(%s)", c.ID)
	}
	defer func() {
		if err0 != nil {
			defer c.unmountVolumes()
		}
	}()

	resolvedPath, absPath := c.getResolvedPath(path)
	lstat, err := os.Lstat(resolvedPath)
	if err != nil {
		return nil, nil, err
	}

	stat = &types.ContainerPathStat{
		Name:  lstat.Name(),
		Path:  absPath,
		Size:  strconv.FormatInt(lstat.Size(), 10),
		Mode:  uint32(lstat.Mode()),
		Mtime: strfmt.DateTime(lstat.ModTime()),
	}
	// TODO: support follow link in container rootfs
	copyInfo, err := archive.CopyInfoSourcePath(resolvedPath, false)
	if err != nil {
		return nil, nil, err
	}
	data, err := archive.TarResource(copyInfo)
	if err != nil {
		return nil, nil, err
	}

	// wait for io finish, then unmount the rootfs
	content = ioutils.NewReadCloserWrapper(data, func() error {
		err := data.Close()
		if !running {
			mgr.detachVolumes(ctx, c, false)
		}
		c.unmountVolumes()
		mgr.Unmount(ctx, c, false, !running)
		c.Unlock()
		return err
	})

	return content, stat, nil
}

// ExtractToDir extracts the given archive at the specified path in the container.
func (mgr *ContainerManager) ExtractToDir(ctx context.Context, name, path string, copyUIDGID, noOverwriteDirNonDir bool, content io.Reader) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}
	c.Lock()
	defer c.Unlock()

	if c.State.Dead {
		return pkgerrors.Errorf("container(%s) has been deleted", c.ID)
	}

	running := c.IsRunningOrPaused()
	err = mgr.Mount(ctx, c, false)
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to mount cid(%s)", c.ID)
	}
	defer mgr.Unmount(ctx, c, false, !running)

	if !running {
		err = mgr.attachVolumes(ctx, c)
		if err != nil {
			return pkgerrors.Wrapf(err, "failed to attachVolumes cid(%s)", c.ID)
		}
		defer mgr.detachVolumes(ctx, c, false)
	}

	err = c.mountVolumes()
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to mountVolumes cid(%s)", c.ID)
	}
	defer c.unmountVolumes()

	resolvedPath, _ := c.getResolvedPath(path)

	lstat, err := os.Lstat(resolvedPath)
	if err != nil {
		return err
	}

	if !lstat.IsDir() {
		return errors.New("can't extract to not dir position")
	}

	// first check if the dir in volume
	inVolume := false
	for _, mp := range c.Mounts {
		if !strings.HasPrefix(path, mp.Destination) {
			continue
		}
		inVolume = true
		if mp.RW {
			break
		}
		return errors.New("can't extract to dir because volume read only")
	}

	if !inVolume && c.HostConfig.ReadonlyRootfs {
		return errors.New("can't extract to dir because rootfs read only")
	}

	// TODO: support copy uid/gid maps
	opts := &archive.TarOptions{
		NoOverwriteDirNonDir: noOverwriteDirNonDir,
	}

	return chrootarchive.Untar(content, resolvedPath, opts)
}

func (c *Container) getResolvedPath(path string) (resolvedPath, absPath string) {
	// consider the given path as an absolute path in the container.
	absPath = path
	if !filepath.IsAbs(absPath) {
		absPath = archive.PreserveTrailingDotOrSeparator(filepath.Join(string(os.PathSeparator), path), path, os.PathSeparator)
	}

	// get the real path on the host
	resolvedPath = filepath.Join(c.BaseFS, absPath)
	resolvedPath = filepath.Clean(resolvedPath)

	return resolvedPath, absPath
}

func (c *Container) mountVolumes() (err0 error) {
	rollbackMounts := make([]string, 0, len(c.Mounts))

	defer func() {
		if err0 != nil {
			for _, dest := range rollbackMounts {
				if err := mount.Unmount(dest); err != nil {
					logrus.Warnf("[mountVolumes:rollback] failed to unmount(%s), err(%v)", dest, err)
				} else {
					logrus.Debugf("[mountVolumes:rollback] unmount(%s)", dest)
				}
			}
		}
	}()

	for _, m := range c.Mounts {
		dest, _ := c.getResolvedPath(m.Destination)

		logrus.Debugf("try to mount volume(source %s -> dest %s", m.Source, dest)

		stat, err := os.Stat(m.Source)
		if err != nil {
			return err
		}

		// /dev /proc /tmp is mounted by default oci spec. Those mount
		// points are mounted by runC. In k8s, the daemonset will mount
		// log into /dev/termination-log. Before mount it, we need to
		// create folder or empty file for the target. It will not
		// impact the running container.
		if err = fileutils.CreateIfNotExists(dest, stat.IsDir()); err != nil {
			return err
		}

		writeMode := "ro"
		if m.RW {
			writeMode = "rw"
		}

		// mountVolumes() seems to be called for temporary mounts
		// outside the container. Soon these will be unmounted with
		// lazy unmount option and given we have mounted the rbind,
		// all the submounts will propagate if these are shared. If
		// daemon is running in host namespace and has / as shared
		// then these unmounts will propagate and unmount original
		// mount as well. So make all these mounts rprivate.
		// Do not use propagation property of volume as that should
		// apply only when mounting happens inside the container.
		opts := strings.Join([]string{"bind", writeMode, "rprivate"}, ",")
		if err := mount.Mount(m.Source, dest, "", opts); err != nil {
			return err
		}

		rollbackMounts = append(rollbackMounts, dest)
	}

	return nil
}

func (c *Container) unmountVolumes() error {
	for _, m := range c.Mounts {
		dest, _ := c.getResolvedPath(m.Destination)

		if err := mount.Unmount(dest); err != nil {
			return err
		}
	}
	return nil
}
