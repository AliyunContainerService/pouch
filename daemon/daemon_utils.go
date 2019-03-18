package daemon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"

	"github.com/containerd/containerd/runtime/linux/runctypes"
	runcoptions "github.com/containerd/containerd/runtime/v2/runc/options"
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
// this solution would be deprecated after shim v1 is deprecated.
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
		if r.Path == "" {
			r.Path = name
		}

		// setup a fake path
		if len(r.RuntimeArgs) != 0 {
			data := fmt.Sprintf("#!/bin/sh\n%s %s $@\n", r.Path, strings.Join(r.RuntimeArgs, " "))
			r.Path = filepath.Join(dir, name)

			if err := ioutil.WriteFile(r.Path, []byte(data), runtimeScriptPerm); err != nil {
				return fmt.Errorf("failed to create runtime script %s: %s", r.Path, err)
			}
		}

		if r.Type == "" {
			r.Type = ctrd.RuntimeTypeV1
		}

		options := getRuntimeOptionsType(r.Type)
		if options != nil {
			// convert general json map to specific options type
			b, err := json.Marshal(r.Options)
			if err != nil {
				return fmt.Errorf("failed to marshal options, runtime: %s: %v", name, err)
			}
			if err := json.Unmarshal(b, options); err != nil {
				return fmt.Errorf("failed to unmarshal to type %+v: %v", options, err)
			}
		}

		r.Options = options

		runtimes[name] = r
	}

	return nil
}

func getRuntimeOptionsType(runtimeType string) interface{} {
	switch runtimeType {
	case
		ctrd.RuntimeTypeV1,
		ctrd.RuntimeTypeV2runscV1,
		ctrd.RuntimeTypeV2kataV2:
		return &runctypes.RuncOptions{}
	case ctrd.RuntimeTypeV2runcV1:
		return &runcoptions.Options{}
	default:
		return nil
	}
}
