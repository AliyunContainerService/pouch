package daemon

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"

	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/snapshots"
	"github.com/sirupsen/logrus"
)

var (
	runtimeDir                    = "runtimes"
	runtimeDirPerm    os.FileMode = 0700
	runtimeScriptPerm os.FileMode = 0700
)

// initialRuntime initializes real runtime path. If runtime.args passed,
// we will make a executable script as a path, or runtime.path is record as path.
// NOTE: containerd not support runtime args directly, so we make executable
// script include runtime path and args as a runtime execute binary.
func initialRuntime(baseDir string, runtimes map[string]types.Runtime) error {
	dir := filepath.Join(baseDir, runtimeDir)

	// remove runtime scripts last generated, since runtime config may changed
	// every time daemon start.
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to clean runtime scripts directory %s: %s", dir, err)
	}

	if err := os.MkdirAll(dir, runtimeDirPerm); err != nil {
		return fmt.Errorf("failed to new runtime scripts directory %s: %s", dir, err)
	}

	// create script for runtime who has args
	for name, r := range runtimes {
		if len(r.RuntimeArgs) == 0 {
			continue
		}

		script := filepath.Join(dir, name)
		if r.Path == "" {
			r.Path = name
		}
		data := fmt.Sprintf("#!/bin/sh\n%s %s $@\n", r.Path, strings.Join(r.RuntimeArgs, " "))

		if err := ioutil.WriteFile(script, []byte(data), runtimeScriptPerm); err != nil {
			return fmt.Errorf("failed to create runtime script %s: %s", script, err)
		}
	}

	return nil
}

// checkSnapshotterValid checks whether the given snapshotter is valid
func checkSnapshotterValid(snapshotter string, ctrdClient ctrd.APIClient) error {
	var (
		driverFound = false
	)

	plugins, err := ctrdClient.Plugins(context.Background(), []string{fmt.Sprintf("type==%s", plugin.SnapshotPlugin)})
	if err != nil {
		logrus.Errorf("failed to get containerd plugins: %v", err)
		return err
	}

	for _, p := range plugins {
		if p.Status != ctrd.PluginStatusOk {
			continue
		}

		if p.ID == snapshotter {
			driverFound = true
			continue
		}

		// check if other snapshotter exists snapshots
		exist, err := checkSnapshotsExist(p.ID, ctrdClient)
		if exist {
			return fmt.Errorf("current snapshotter driver is %s, cannot change to %s", p.ID, snapshotter)
		}

		if err != nil {
			return fmt.Errorf("failed to check snapshotter driver %s: %v", p.ID, err)
		}
	}

	if !driverFound {
		return fmt.Errorf("containerd not support snapshotter driver %s", snapshotter)
	}

	return nil
}

func checkSnapshotsExist(snapshotter string, ctrdClient ctrd.APIClient) (existSnapshot bool, err error) {
	fn := func(c context.Context, s snapshots.Info) error {
		existSnapshot = true
		return nil
	}

	err = ctrdClient.WalkSnapshot(context.Background(), snapshotter, fn)
	return
}
