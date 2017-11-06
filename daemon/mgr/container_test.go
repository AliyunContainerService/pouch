package mgr

import (
	"fmt"
	"testing"

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
		arr, err := checkBind(p.bind)
		if p.err {
			assert.Equal(err, p.expectErr)
		} else {
			assert.NoError(err, p.expectErr)
			assert.Equal(len(arr), p.len)
		}
	}
}
