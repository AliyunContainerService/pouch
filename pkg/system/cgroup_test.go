package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCgroupEnable(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		exist bool
		fs    []string
	}

	tmpDir, err := ioutil.TempDir("", "test-cgroup-enable")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	existFile := filepath.Join(tmpDir, "exist")
	fd, err := os.Create(existFile)
	assert.NoError(err)
	fd.Close()

	for _, tc := range []tCase{
		{
			exist: false,
			fs:    []string{tmpDir, "foo"},
		},
		{
			exist: false,
			fs:    []string{tmpDir, "foo", "bar"},
		},
		{
			exist: true,
			fs:    []string{tmpDir, "exist"},
		},
	} {
		assert.Equal(tc.exist, isCgroupEnable(tc.fs...))
	}
}

func TestGetCgroupRootMount(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		cgroupMount string
		data        string
	}

	tmpDir, err := ioutil.TempDir("", "test-cgroup-enable")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)
	file := filepath.Join(tmpDir, "mountinfo")

	path := filepath.Join(tmpDir, "cgroup")
	fd, err := os.Create(path)
	fd.Close()
	assert.NoError(err)

	for _, tc := range []tCase{
		{
			cgroupMount: "",
			data:        "",
		},
		{
			cgroupMount: "",
			data:        "a b\n a b",
		},
		{
			cgroupMount: "",
			data:        "18 58 0:17 / /sys rw,relatime shared:6 - sysfs sysfs rw",
		},
		{
			cgroupMount: tmpDir,
			data:        fmt.Sprintf("a a a a %s a a - cgroup a a\n a b c", path),
		},
	} {
		err := ioutil.WriteFile(file, []byte(tc.data), 0644)
		assert.NoError(err)
		assert.Equal(tc.cgroupMount, getCgroupRootMount(file))
	}
}
