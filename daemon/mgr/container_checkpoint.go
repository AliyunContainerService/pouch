package mgr

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
)

var (
	checkpointConfigPath             = "config.json"
	checkpointConfigPerm os.FileMode = 0700
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
func (mgr *ContainerManager) CreateCheckpoint(ctx context.Context, name string, options *types.CheckpointCreateOptions) (err0 error) {
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
	defer func() {
		if err0 != nil {
			os.RemoveAll(dir)
		}
	}()

	if err := mgr.Client.CreateCheckpoint(ctx, c.ID, dir, options.Exit); err != nil {
		return err
	}

	return writeCheckpointConfig(filepath.Join(dir, checkpointConfigPath), c.ID, options.CheckpointID)
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
		path := filepath.Join(dir, checkpoint.Name(), checkpointConfigPath)
		if config, err := readCheckpointConfig(path); err == nil &&
			config != nil && config.ContainerID == c.ID {
			cpList = append(cpList, config.CheckpointName)
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

func writeCheckpointConfig(path, container, checkpoint string) error {
	config := &types.Checkpoint{
		ContainerID:    container,
		CheckpointName: checkpoint,
	}

	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, raw, checkpointConfigPerm)
}

func readCheckpointConfig(path string) (*types.Checkpoint, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	config := &types.Checkpoint{}
	if err = json.Unmarshal(raw, config); err != nil {
		return nil, err
	}

	return config, err
}
