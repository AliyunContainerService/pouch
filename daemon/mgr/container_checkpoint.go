package mgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
)

// getCheckpointDir gets container checkpoint directory.
func (mgr *ContainerManager) getCheckpointDir(container, prefixDir, checkpointID string, create bool) (string, error) {
	if prefixDir == "" {
		prefixDir = filepath.Join(mgr.Store.Path(container), "checkpoint")
	}
	checkpointDir := filepath.Join(prefixDir, checkpointID)

	stat, err := os.Stat(checkpointDir)
	if create {
		switch {
		case err != nil && os.IsNotExist(err):
			return checkpointDir, os.MkdirAll(checkpointDir, 0700)
		case err != nil:
			return "", fmt.Errorf("failed to create checkpoint %s", checkpointID)
		case !stat.IsDir():
			return "", fmt.Errorf("checkpoint %s exist but not directory", checkpointID)
		default:
			return "", fmt.Errorf("checkpoint %s is already exist", checkpointID)
		}
	}

	switch {
	case err == nil && stat.IsDir():
		return checkpointDir, nil
	case err == nil:
		return "", fmt.Errorf("checkpoint %s exist but not directory", checkpointID)
	}

	return "", fmt.Errorf("checkpoint %s is not exist for container %s", checkpointID, container)
}

// CreateCheckpoint creates a checkpoint from a running container
func (mgr *ContainerManager) CreateCheckpoint(ctx context.Context, name string, options *types.CheckpointCreateOptions) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	if c.State.Status != types.StatusRunning {
		return fmt.Errorf("can not checkpoint from a %s container", c.State.Status)
	}

	if c.Config.Tty {
		return fmt.Errorf("checkpoint not support on containers with tty")
	}

	dir, err := mgr.getCheckpointDir(c.ID, options.CheckpointDir, options.CheckpointID, true)
	if err != nil {
		return err
	}

	if err := mgr.Client.CreateCheckpoint(ctx, c.ID, dir, options.Exit); err != nil {
		return err
	}

	return nil
}

// ListCheckpoint lists checkpoints from a container
func (mgr *ContainerManager) ListCheckpoint(ctx context.Context, name string, options *types.CheckpointListOptions) ([]string, error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, err
	}

	dir, err := mgr.getCheckpointDir(c.ID, options.CheckpointDir, "", false)
	if err != nil {
		// if error returns, it means no checkpoint has been created under
		// the specified checkpoint directory, return nil is ok.
		return nil, nil
	}

	checkpoints, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	cpList := make([]string, 0)
	for _, checkpoint := range checkpoints {
		cp := filepath.Join(dir, checkpoint.Name())
		if stat, err := os.Stat(cp); err == nil && stat.IsDir() {
			cpList = append(cpList, checkpoint.Name())

		}
	}

	return cpList, nil
}

// DeleteCheckpoint deletes a checkpoint from a container
func (mgr *ContainerManager) DeleteCheckpoint(ctx context.Context, name string, options *types.CheckpointDeleteOptions) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	dir, err := mgr.getCheckpointDir(c.ID, options.CheckpointDir, options.CheckpointID, false)
	if err != nil {
		return err
	}

	return os.RemoveAll(dir)
}
