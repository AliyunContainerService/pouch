package mgr

import (
	"testing"

	"github.com/alibaba/pouch/daemon/config"
	"github.com/stretchr/testify/assert"
)

func testAddRegistry(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		repo   string
		expect string
	}

	var (
		registry = "testRegistry"
		ns       = "testNS"
	)
	// set default registry and namespace
	imgr, err := NewImageManager(&config.Config{
		DefaultRegistry:   registry,
		DefaultRegistryNS: ns,
	}, nil)
	assert.NoError(err)

	for _, tc := range []tCase{
		{
			repo:   "docker.io/library/busybox",
			expect: "docker.io/library/busybox",
		},
		{
			repo:   "library/busybox",
			expect: registry + "/library/busybox",
		},
		{
			repo:   "127.0.0.1:5000/bar",
			expect: "127.0.0.1:5000/bar",
		},
		{
			repo:   "0.0.0.0/bar",
			expect: "0.0.0.0/bar",
		},
		{
			repo:   "registry.com/bar",
			expect: "registry.com/bar",
		},
		{
			repo:   "bar",
			expect: registry + "/" + ns + "/bar",
		},
		{
			repo:   "foo/bar",
			expect: registry + "/foo/bar",
		},
	} {
		output := imgr.addRegistry(tc.repo)
		assert.Equal(tc.expect, output)
	}
}
