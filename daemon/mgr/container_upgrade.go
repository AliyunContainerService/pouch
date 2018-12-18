package mgr

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Upgrade a container with new image and args. when upgrade a container,
// we only support specify cmd and entrypoint. if you want to change other
// parameters of the container, you should think about the update API first.
func (mgr *ContainerManager) Upgrade(ctx context.Context, name string, config *types.ContainerUpgradeConfig) error {
	var err error
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	var (
		needRollback  = false
		oldConfig     = *c.Config
		oldHostconfig = *c.HostConfig
		oldImage      = c.Image
		oldSnapID     = c.SnapshotKey()
		IsRunning     = false
	)

	// use err to determine if we should recover old container configure.
	defer func() {
		if err == nil {
			return
		}

		c.Lock()
		// recover old container config
		c.Config = &oldConfig
		c.HostConfig = &oldHostconfig
		c.Image = oldImage
		c.SnapshotID = oldSnapID
		c.Unlock()

		// even if the err is not nil, we may still no need to rollback the container
		if !needRollback {
			return
		}

		if err := mgr.start(ctx, c, &types.ContainerStartOptions{}); err != nil {
			logrus.Errorf("failed to rollback upgrade action: %s", err.Error())
			if err := mgr.markStoppedAndRelease(c, nil); err != nil {
				logrus.Errorf("failed to mark container %s stop status: %s", c.ID, err.Error())
			}
		}
	}()

	// merge image config to container config
	err = mgr.mergeImageConfigForUpgrade(ctx, c, config)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade container")
	}

	// if the container is running, we need first stop it.
	if c.IsRunning() {
		IsRunning = true
		err = mgr.stop(ctx, c, 10)
		if err != nil {
			return errors.Wrapf(err, "failed to stop container %s when upgrade", c.Key())
		}
		needRollback = true
	}

	// prepare new snapshot for the new container
	newSnapID, err := mgr.prepareSnapshotForUpgrade(ctx, c.Key(), c.SnapshotKey(), config.Image)
	if err != nil {
		return err
	}
	c.SetSnapshotID(newSnapID)

	// initialize container storage config before container started
	err = mgr.initContainerStorage(ctx, c)
	if err != nil {
		return errors.Wrapf(err, "failed to init container storage, id: (%s)", c.Key())
	}

	// If container is running, we also should start the container
	// after recreate it.
	if IsRunning {
		err = mgr.start(ctx, c, &types.ContainerStartOptions{})
		if err != nil {
			if err := mgr.Client.RemoveSnapshot(ctx, newSnapID); err != nil {
				logrus.Errorf("failed to remove snapshot %s: %v", newSnapID, err)
			}
		}

		return errors.Wrap(err, "failed to create new container")
	}

	// Upgrade success, remove snapshot of old container
	if err := mgr.Client.RemoveSnapshot(ctx, oldSnapID); err != nil {
		// TODO(ziren): remove old snapshot failed, may cause dirty data
		logrus.Errorf("failed to remove snapshot %s: %v", oldSnapID, err)
	}

	// Upgrade succeeded, refresh the cache
	mgr.cache.Put(c.ID, c)

	// Works fine, store new container info to disk.
	if err := c.Write(mgr.Store); err != nil {
		logrus.Errorf("failed to update container %s in meta store: %v", c.ID, err)
		return err
	}

	return nil
}

func (mgr *ContainerManager) prepareContainerEntrypointForUpgrade(ctx context.Context, c *Container, config *types.ContainerUpgradeConfig) error {
	// Firstly, try to use the entrypoint specified by ContainerUpgradeConfig
	if len(config.Entrypoint) > 0 || len(config.Cmd) > 0 {
		return nil
	}

	// Secondly, try to use the entrypoint of the old container.
	// because of the entrypoints of old container's CreateConfig and old image being merged,
	// so we cannot decide which config that the old container's entrypoint belongs to, so just check if the old
	// container's entrypoint is different with the old image.
	c.Lock()
	defer c.Unlock()

	oldImgConfig, err := mgr.ImageMgr.GetOCIImageConfig(ctx, c.Config.Image)
	if err != nil {
		return err
	}

	// if the entrypoints of old container and the old image is empty, we should use to CMD of old container,
	// else if entrypoints are different, we use the CMD of old container.
	if (c.Config.Entrypoint == nil && oldImgConfig.Entrypoint == nil) || !utils.StringSliceEqual(c.Config.Entrypoint, oldImgConfig.Entrypoint) {
		config.Entrypoint = c.Config.Entrypoint
		if len(config.Cmd) == 0 {
			config.Cmd = c.Config.Cmd
		}

		return nil
	}

	// Thirdly, just use the entrypoint of the new image
	newImgConfig, err := mgr.ImageMgr.GetOCIImageConfig(ctx, config.Image)
	if err != nil {
		return err
	}
	config.Entrypoint = newImgConfig.Entrypoint
	if len(config.Cmd) == 0 {
		config.Cmd = newImgConfig.Cmd
	}
	return nil
}

func (mgr *ContainerManager) prepareSnapshotForUpgrade(ctx context.Context, cID, oldSnapID, image string) (string, error) {
	newSnapID := ""
	// get a ID for the new snapshot
	for {
		newSnapID = utils.RandString(8, cID, "")
		if newSnapID != oldSnapID {
			break
		}
	}

	// create a snapshot with image for new container.
	if err := mgr.Client.CreateSnapshot(ctx, newSnapID, image, nil); err != nil {
		return "", errors.Wrap(err, "failed to create snapshot")
	}

	return newSnapID, nil
}

func (mgr *ContainerManager) mergeImageConfigForUpgrade(ctx context.Context, c *Container, config *types.ContainerUpgradeConfig) error {
	// check the image existed or not, and convert image id to image ref
	imgID, _, primaryRef, err := mgr.ImageMgr.CheckReference(ctx, config.Image)
	if err != nil {
		return errors.Wrap(err, "failed to get image")
	}

	config.Image = primaryRef.String()
	// Nothing changed, no need upgrade.
	if config.Image == c.Config.Image {
		return fmt.Errorf("failed to upgrade container: image not changed")
	}

	// prepare entrypoint for the new container of upgrade
	if err := mgr.prepareContainerEntrypointForUpgrade(ctx, c, config); err != nil {
		return errors.Wrap(err, "failed to check entrypoint for container upgrade")
	}

	// merge image's config into the new container of upgrade
	if err := c.merge(func() (ocispec.ImageConfig, error) {
		return mgr.ImageMgr.GetOCIImageConfig(ctx, config.Image)
	}); err != nil {
		return errors.Wrap(err, "failed to merge image config when upgrade container")
	}

	// set image and entrypoint for new container.
	c.Lock()
	c.Image = imgID.String()
	c.Config.Image = config.Image
	c.Config.Entrypoint = config.Entrypoint
	c.Config.Cmd = config.Cmd
	c.Unlock()

	return nil
}
