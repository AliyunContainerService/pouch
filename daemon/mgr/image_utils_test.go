package mgr

import (
	"testing"

	"github.com/alibaba/pouch/pkg/reference"
	"github.com/stretchr/testify/assert"
)

func TestAddDefaultRegistry(t *testing.T) {
	defaultRegistry, defaultNamespace := "pouch.io", "library"

	for _, tc := range []struct {
		repo   string
		expect string
	}{
		{
			repo:   "docker.io/library/busybox",
			expect: "docker.io/library/busybox",
		}, {
			repo:   "library/busybox",
			expect: defaultRegistry + "/library/busybox",
		}, {
			repo:   "127.0.0.1:5000/bar",
			expect: "127.0.0.1:5000/bar",
		},
		{
			repo:   "0.0.0.0/bar",
			expect: "0.0.0.0/bar",
		}, {
			repo:   "registry.com/bar",
			expect: "registry.com/bar",
		}, {
			repo:   "bar",
			expect: defaultRegistry + "/" + defaultNamespace + "/" + "bar",
		}, {
			repo:   "busybox@sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
			expect: defaultRegistry + "/" + defaultNamespace + "/" + "busybox@sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
		}, {
			repo:   "foo/bar",
			expect: defaultRegistry + "/foo/bar",
		}, {
			repo:   "foo/bar@sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
			expect: defaultRegistry + "/foo/bar@sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
		},
	} {
		assert.Equal(t, addDefaultRegistryIfMissing(tc.repo, defaultRegistry, defaultNamespace), tc.expect)
	}
}

func TestUniqueLocatorReference(t *testing.T) {
	for _, tc := range []struct {
		refs   []string
		expect bool
	}{
		{
			refs: []string{
				"docker.io/busybox:1.28",
				"docker.io/busybox:latest",
				"docker.io/busybox@sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
			},
			expect: true,
		}, {
			refs: []string{
				"library/busybox:latest",
				"localhost:5000/busybox:1.25",
			},
			expect: false,
		}, {
			refs:   []string{"busybox:1.28"},
			expect: true,
		}, {
			refs:   []string{},
			expect: true,
		},
	} {
		refs := make([]reference.Named, 0, len(tc.refs))
		for _, ref := range tc.refs {
			namedRef, err := reference.Parse(ref)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", ref, err)
			}
			refs = append(refs, namedRef)
		}
		assert.Equal(t, uniqueLocatorReference(refs), tc.expect)
	}
}
