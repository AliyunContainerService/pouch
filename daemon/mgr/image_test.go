package mgr

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestImageCache(t *testing.T) {
	images := []types.ImageInfo{
		{
			RepoDigests: []string{"docker.io/library/nginx@sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3"},
			RepoTags:    []string{"docker.io/library/nginx:alpine"},
			ID:          "sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3",
		},
		{
			RepoDigests: []string{"reg.docker.alibaba-inc.com/base/hello-world@sha256:2075ac87b043415d35bb6351b4a59df19b8ad154e578f7048335feeb02d0f759"},
			RepoTags:    []string{"reg.docker.alibaba-inc.com/base/hello-world:latest"},
			ID:          "sha256:2075ac87b043415d35bb6351b4a59df19b8ad154e578f7048335feeb02d0f759",
		},
		{
			RepoDigests: []string{"reg.docker.alibaba-inc.com/base/hello-world@sha256:ded83fbb30d5fad804784215bc454c3844653cb0a907a512d44e25429507e415s"},
			RepoTags:    []string{"reg.docker.alibaba-inc.com/busybox:latest"},
			ID:          "sha256:ded83fbb30d5fad804784215bc454c3844653cb0a907a512d44e25429507e415s",
		},
		{
			RepoDigests: []string{"reg.docker.alibaba-inc.com/base/hello-world@sha256:6591a3cb89fec995f299fa52c65e56aa33c779fd965060cf3b7759cd4b346aac"},
			RepoTags:    []string{"reg.docker.alibaba-inc.com/fedora:21"},
			ID:          "sha256:6591a3cb89fec995f299fa52c65e56aa33c779fd965060cf3b7759cd4b346aac",
		},
		{
			RepoDigests: []string{"reg.docker.alibaba-inc.com/base/hello-world@sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928"},
			RepoTags:    []string{"reg.docker.alibaba-inc.com/mysql:5.6.23"},
			ID:          "sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928",
		},
	}

	cache := newImageCache()

	for i := range images {
		cache.put(&images[i])
	}

	if i, err := cache.get("dc5f67a48da7"); err != nil {
		t.Fatal(err)
	} else if i.RepoTags[0] != "docker.io/library/nginx:alpine" {
		t.Errorf("get error: %s", i.RepoTags[0])
	}

	if i, err := cache.get("ded83fbb30d5"); err != nil {
		t.Fatal(err)
	} else if i.RepoTags[0] != "reg.docker.alibaba-inc.com/busybox:latest" {
		t.Errorf("get error: %s", i.RepoTags[0])
	}

	if i, err := cache.get("reg.docker.alibaba-inc.com/mysql:5.6.23"); err != nil {
		t.Fatal(err)
	} else if i.RepoTags[0] != "reg.docker.alibaba-inc.com/mysql:5.6.23" {
		t.Errorf("get error: %s", i.RepoTags[0])
	}

	cache.remove(&types.ImageInfo{
		RepoDigests: []string{"reg.docker.alibaba-inc.com/mysql@sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928"},
		RepoTags:    []string{"reg.docker.alibaba-inc.com/mysql:5.6.23"},
		ID:          "sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928",
	})

	if i, err := cache.get("reg.docker.alibaba-inc.com/mysql:5.6.23"); err == nil {
		t.Errorf("remove failed: %s", i.RepoTags[0])
	}
	if i, err := cache.get("03b17f82af79"); err == nil {
		t.Errorf("remove failed: %s", i.RepoTags[0])
	}
}
