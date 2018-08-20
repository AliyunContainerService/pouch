package daemon

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestInitialRuntime(t *testing.T) {
	assert := assert.New(t)
	tmpDir, err := ioutil.TempDir("", "runtime-path")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	runtimeDir = "runtimes"

	for _, tc := range []struct {
		runtimes map[string]types.Runtime
		rname    string
		rpath    string
		filedata string
	}{
		{
			runtimes: map[string]types.Runtime{
				"c": {
					Path: "/foo/bar/c",
					RuntimeArgs: []string{
						"--foo=foo",
						"--bar=bar",
					},
				},
			},
			rname:    "c",
			rpath:    filepath.Join(tmpDir, runtimeDir, "c"),
			filedata: "#!/bin/sh\n/foo/bar/c --foo=foo --bar=bar $@\n",
		},
		{
			runtimes: map[string]types.Runtime{
				"d": {
					RuntimeArgs: []string{
						"--foo=foo",
						"--bar=bar",
					},
				},
			},
			rname:    "d",
			rpath:    filepath.Join(tmpDir, runtimeDir, "d"),
			filedata: "#!/bin/sh\nd --foo=foo --bar=bar $@\n",
		},
	} {
		err = initialRuntime(tmpDir, tc.runtimes)
		assert.NoError(err)
		if tc.filedata != "" {
			if _, err := os.Stat(tc.rpath); err != nil {
				t.Fatalf("%s should exist", tc.rpath)
			}
			data, err := ioutil.ReadFile(tc.rpath)
			assert.NoError(err)
			assert.Equal(tc.filedata, string(data))
		}
	}
}
