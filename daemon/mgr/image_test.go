package mgr

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestImageCache(t *testing.T) {
	images := []types.Image{
		{
			Digest: "sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3",
			Name:   "docker.io/library/nginx:alpine",
			ID:     "dc5f67a48da7",
		},
		{
			Digest: "sha256:2075ac87b043415d35bb6351b4a59df19b8ad154e578f7048335feeb02d0f759",
			Name:   "reg.docker.alibaba-inc.com/base/hello-world:latest",
			ID:     "2075ac87b043",
		},
		{
			Digest: "sha256:ded83fbb30d5fad804784215bc454c3844653cb0a907a512d44e25429507e415",
			Name:   "reg.docker.alibaba-inc.com/busybox:latest",
			ID:     "ded83fbb30d5",
		},
		{
			Digest: "sha256:6591a3cb89fec995f299fa52c65e56aa33c779fd965060cf3b7759cd4b346aac",
			Name:   "reg.docker.alibaba-inc.com/fedora:21",
			ID:     "6591a3cb89fe",
		},
		{
			Digest: "sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928",
			Name:   "reg.docker.alibaba-inc.com/mysql:5.6.23",
			ID:     "03b17f82af79",
		},
	}

	cache := newImageCache()

	for i := range images {
		cache.put(&images[i])
	}

	if i, err := cache.get("dc5f67a48da7"); err != nil {
		t.Fatal(err)
	} else if i.Name != "docker.io/library/nginx:alpine" {
		t.Errorf("get error: %s", i.Name)
	}

	if i, err := cache.get("ded83fbb30d5"); err != nil {
		t.Fatal(err)
	} else if i.Name != "reg.docker.alibaba-inc.com/busybox:latest" {
		t.Errorf("get error: %s", i.Name)
	}

	if i, err := cache.get("reg.docker.alibaba-inc.com/mysql:5.6.23"); err != nil {
		t.Fatal(err)
	} else if i.Name != "reg.docker.alibaba-inc.com/mysql:5.6.23" {
		t.Errorf("get error: %s", i.Name)
	}

	cache.remove(&types.Image{
		Digest: "sha256:03b17f82af79a338571410df9b26670c160175eda81d6e9fd62e3fda6b728928",
		Name:   "reg.docker.alibaba-inc.com/mysql:5.6.23",
		ID:     "03b17f82af79",
	})

	if i, err := cache.get("reg.docker.alibaba-inc.com/mysql:5.6.23"); err == nil {
		t.Errorf("remove failed: %s", i.Name)
	}
	if i, err := cache.get("03b17f82af79"); err == nil {
		t.Errorf("remove failed: %s", i.Name)
	}
}
