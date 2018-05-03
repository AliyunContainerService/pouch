package mgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// containerdImageToOciImage returns the oci image spec.
func containerdImageToOciImage(ctx context.Context, img containerd.Image) (ocispec.Image, error) {
	var ociImage ocispec.Image

	cfg, err := img.Config(ctx)
	if err != nil {
		return ocispec.Image{}, err
	}

	switch cfg.MediaType {
	case ocispec.MediaTypeImageConfig, images.MediaTypeDockerSchema2Config:
		data, err := content.ReadBlob(ctx, img.ContentStore(), cfg.Digest)
		if err != nil {
			return ocispec.Image{}, err
		}

		if err := json.Unmarshal(data, &ociImage); err != nil {
			return ocispec.Image{}, err
		}
	default:
		return ocispec.Image{}, fmt.Errorf("unknown image config media type %s", cfg.MediaType)
	}

	return ociImage, nil
}

// getImageInfoConfigFromOciImage returns config of ImageConfig from oci image.
func getImageInfoConfigFromOciImage(img ocispec.Image) *types.ContainerConfig {
	volumes := make(map[string]interface{})
	for k, obj := range img.Config.Volumes {
		volumes[k] = obj
	}

	return &types.ContainerConfig{
		User:       img.Config.User,
		Env:        img.Config.Env,
		Entrypoint: img.Config.Entrypoint,
		Cmd:        img.Config.Cmd,
		WorkingDir: img.Config.WorkingDir,
		Labels:     img.Config.Labels,
		StopSignal: img.Config.StopSignal,
		Volumes:    volumes,
	}
}

func digestSliceToStringSlice(from []digest.Digest) []string {
	to := make([]string, 0, len(from))
	for _, f := range from {
		to = append(to, f.String())
	}
	return to
}

// addDefaultRegistryIfMissing will add default registry and namespace if missing.
func addDefaultRegistryIfMissing(ref string, defaultRegistry, defaultNamespace string) string {
	var (
		registry  string
		remainder string
	)

	idx := strings.IndexRune(ref, '/')
	if idx == -1 || !strings.ContainsAny(ref[:idx], ".:") {
		registry, remainder = defaultRegistry, ref
	} else {
		registry, remainder = ref[:idx], ref[idx+1:]
	}

	if registry == defaultRegistry && !strings.ContainsAny(remainder, "/") {
		remainder = defaultNamespace + "/" + remainder
	}
	return registry + "/" + remainder
}

// uniqueLocatorReference checks the references have the same locator.
//
// For example,
//
//	A. localhost:5000/busybox:latest
//	B. localhost:5000/busybox@sha256:xxxx
//	C. docker.io/busybox:latest
//
// Both A and B has the same locator, but the C doesn't.
func uniqueLocatorReference(refs []reference.Named) bool {
	var locator string
	for _, ref := range refs {
		if locator == "" {
			locator = ref.Name()
			continue
		}

		if locator != ref.Name() {
			return false
		}
	}
	return true
}
