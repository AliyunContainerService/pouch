package opts

import (
	"fmt"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestCheckBind(t *testing.T) {
	assert := assert.New(t)

	type parsed struct {
		bind      string
		len       int
		err       bool
		expectErr error
	}

	parseds := []parsed{
		{bind: "volume-test:/mnt", len: 2, err: false, expectErr: fmt.Errorf("")},
		{bind: "volume-test:/mnt:rw", len: 3, err: false, expectErr: fmt.Errorf("")},
		{bind: "/mnt", len: 1, err: false, expectErr: fmt.Errorf("")},
		{bind: ":/mnt:rw", len: 3, err: false, expectErr: fmt.Errorf(":/mnt:rw")},
		{bind: "volume-test:/mnt:/mnt:rw", len: 4, err: true, expectErr: fmt.Errorf("unknown volume bind: volume-test:/mnt:/mnt:rw")},
		{bind: "", len: 0, err: true, expectErr: fmt.Errorf("unknown volume bind: ")},
		{bind: "volume-test::rw", len: 3, err: true, expectErr: fmt.Errorf("unknown volume bind: volume-test::rw")},
		{bind: "volume-test", len: 1, err: true, expectErr: fmt.Errorf("invalid bind path: volume-test")},
		{bind: ":mnt:rw", len: 3, err: true, expectErr: fmt.Errorf("invalid bind path: mnt")},
	}

	for _, p := range parseds {
		arr, err := CheckBind(p.bind)
		if p.err {
			assert.Equal(err, p.expectErr)
		} else {
			assert.NoError(err, p.expectErr)
			assert.Equal(len(arr), p.len)
		}
	}
}

func TestParseBindMode(t *testing.T) {
	assert := assert.New(t)

	type parsed struct {
		mode             string
		expectMountPoint *types.MountPoint
		err              bool
		expectErr        error
	}

	parseds := []parsed{
		{mode: "dr", expectMountPoint: &types.MountPoint{Mode: "dr", RW: true, CopyData: true}, err: false, expectErr: nil},
		{mode: "nocopy", expectMountPoint: &types.MountPoint{Mode: "nocopy", RW: true, CopyData: false}, err: false, expectErr: nil},
		{mode: "ro", expectMountPoint: &types.MountPoint{Mode: "ro", RW: false, CopyData: true}, err: false, expectErr: nil},
		{mode: "", expectMountPoint: &types.MountPoint{Mode: "", RW: true, CopyData: true}, err: false, expectErr: nil},
		{mode: "dr,rr", err: true, expectErr: fmt.Errorf("invalid bind mode: dr,rr")},
		{mode: "unknown", err: true, expectErr: fmt.Errorf("unknown bind mode: unknown")},
		{mode: "rw", expectMountPoint: &types.MountPoint{Mode: "rw", RW: true, CopyData: true}, err: false, expectErr: nil},
		{mode: "z,Z", expectMountPoint: &types.MountPoint{Mode: "z,Z", RW: true, CopyData: true}, err: false, expectErr: nil},
	}

	for _, p := range parseds {
		mp := &types.MountPoint{}
		err := ParseBindMode(mp, p.mode)
		if p.err {
			assert.Equal(err, p.expectErr)
		} else {
			assert.NoError(err, p.expectErr)
			assert.Equal(p.expectMountPoint.Mode, mp.Mode)
			assert.Equal(p.expectMountPoint.RW, mp.RW)
			assert.Equal(p.expectMountPoint.CopyData, mp.CopyData)
		}
	}
}

func TestParseVolumesFrom(t *testing.T) {
	assert := assert.New(t)

	type parsed struct {
		volumesFrom string
		expectID    string
		expectMode  string
		err         bool
		expectErr   error
	}

	parseds := []parsed{
		{volumesFrom: "123456789", expectID: "123456789", expectMode: "", err: false, expectErr: nil},
		{volumesFrom: "123456789:nocopy", expectID: "123456789", expectMode: "nocopy", err: false, expectErr: nil},
		{volumesFrom: "123456789:", expectID: "123456789", expectMode: "", err: false, expectErr: nil},
		{volumesFrom: "", expectID: "", expectMode: "", err: true, expectErr: fmt.Errorf("invalid argument volumes-from")},
		{volumesFrom: ":", expectID: "", expectMode: "", err: true, expectErr: fmt.Errorf("failed to parse container's id")},
	}

	for _, p := range parseds {
		containerID, mode, err := ParseVolumesFrom(p.volumesFrom)
		if p.err {
			assert.Equal(err, p.expectErr)
		} else {
			assert.NoError(err, p.expectErr)
			assert.Equal(p.expectID, containerID)
			assert.Equal(p.expectMode, mode)
		}
	}
}
