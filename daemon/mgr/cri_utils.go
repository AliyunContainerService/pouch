package mgr

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/go-openapi/strfmt"

	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

func parseUint32(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func toCriTimestamp(t string) (int64, error) {
	if t == "" {
		return 0, nil
	}

	result, err := time.Parse(utils.TimeLayout, t)
	if err != nil {
		return 0, err
	}

	return result.UnixNano(), nil
}

// generateEnvList converts KeyValue list to a list of strings, in the form of
// '<key>=<value>', which can be understood by pouch.
func generateEnvList(envs []*runtime.KeyValue) (result []string) {
	for _, env := range envs {
		result = append(result, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}
	return
}

func makeLabels(labels, annotations map[string]string) map[string]string {
	m := make(map[string]string)

	for k, v := range labels {
		m[k] = v
	}
	for k, v := range annotations {
		// Use prefix to distinguish between annotations and labels.
		m[fmt.Sprintf("%s%s", annotationPrefix, k)] = v
	}

	return m
}

// extractLabels converts raw pouch labels to the CRI labels and annotations.
func extractLabels(input map[string]string) (map[string]string, map[string]string) {
	labels := make(map[string]string)
	annotations := make(map[string]string)
	for k, v := range input {
		// Check if the key is used internally by the cri manager.
		internal := false
		for _, internalKey := range []string{
			containerTypeLabelKey,
			sandboxIDLabelKey,
		} {
			if k == internalKey {
				internal = true
				break
			}
		}
		if internal {
			continue
		}

		// Check if the label should be treated as an annotation.
		if strings.HasPrefix(k, annotationPrefix) {
			annotations[strings.TrimPrefix(k, annotationPrefix)] = v
			continue
		}
		labels[k] = v
	}

	return labels, annotations
}

// Sandbox related tool functions.

// makeSandboxName generates sandbox name from sandbox metadata. The name
// generated is unique as long as sandbox metadata is unique.
func makeSandboxName(c *runtime.PodSandboxConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		sandboxContainerName,                  // 1
		c.Metadata.Name,                       // 2
		c.Metadata.Namespace,                  // 3
		c.Metadata.Uid,                        // 4
		fmt.Sprintf("%d", c.Metadata.Attempt), // 5
	}, nameDelimiter)
}

func parseSandboxName(name string) (*runtime.PodSandboxMetadata, error) {
	format := fmt.Sprintf("%s_%s_${sandbox name}_${sandbox namespace}_${sandbox uid}_${sandbox attempt}", kubePrefix, sandboxContainerName)

	parts := strings.Split(name, nameDelimiter)
	if len(parts) != 6 {
		return nil, fmt.Errorf("failed to parse sandbox name: %q, which should be %s", name, format)
	}
	if parts[0] != kubePrefix {
		return nil, fmt.Errorf("sandbox container is not managed by kubernetes: %q", name)
	}

	attempt, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("failed to parse the attempt times in sandbox name: %q: %v", name, err)
	}

	return &runtime.PodSandboxMetadata{
		Name:      parts[2],
		Namespace: parts[3],
		Uid:       parts[4],
		Attempt:   attempt,
	}, nil
}

// makeSandboxPouchConfig returns apitypes.ContainerCreateConfig based on runtimeapi.PodSandboxConfig.
func makeSandboxPouchConfig(config *runtime.PodSandboxConfig, image string) (*apitypes.ContainerCreateConfig, error) {
	// Merge annotations and labels because pouch supports only labels.
	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply a label to distinguish sandboxes from regular containers.
	labels[containerTypeLabelKey] = containerTypeLabelSandbox

	hc := &apitypes.HostConfig{}
	createConfig := &apitypes.ContainerCreateConfig{
		ContainerConfig: apitypes.ContainerConfig{
			Hostname: strfmt.Hostname(config.Hostname),
			Image:    image,
			Labels:   labels,
		},
		HostConfig:       hc,
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}

	return createConfig, nil
}

func toCriSandboxState(status apitypes.Status) runtime.PodSandboxState {
	switch status {
	case apitypes.StatusRunning:
		return runtime.PodSandboxState_SANDBOX_READY
	default:
		return runtime.PodSandboxState_SANDBOX_NOTREADY
	}
}

func toCriSandbox(c *ContainerMeta) (*runtime.PodSandbox, error) {
	state := toCriSandboxState(c.State.Status)
	metadata, err := parseSandboxName(c.Name)
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Config.Labels)
	return &runtime.PodSandbox{
		Id:       c.ID,
		Metadata: metadata,
		State:    state,
		// TODO: fill "CreatedAt" when it is appropriate.
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

func filterCRISandboxes(sandboxes []*runtime.PodSandbox, filter *runtime.PodSandboxFilter) []*runtime.PodSandbox {
	if filter == nil {
		return sandboxes
	}

	filtered := []*runtime.PodSandbox{}
	for _, s := range sandboxes {
		if filter.GetId() != "" && filter.GetId() != s.Id {
			continue
		}
		if filter.GetState() != nil && filter.GetState().GetState() != s.State {
			continue
		}
		if filter.GetLabelSelector() != nil {
			match := true
			for k, v := range filter.GetLabelSelector() {
				value, ok := s.Labels[k]
				if !ok || v != value {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, s)
	}

	return filtered
}

// Container related tool functions.

func makeContainerName(s *runtime.PodSandboxConfig, c *runtime.ContainerConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		c.Metadata.Name,                       // 1
		s.Metadata.Name,                       // 2: sandbox name
		s.Metadata.Namespace,                  // 3: sandbox namespace
		s.Metadata.Uid,                        // 4: sandbox uid
		fmt.Sprintf("%d", c.Metadata.Attempt), // 5
	}, nameDelimiter)
}

func parseContainerName(name string) (*runtime.ContainerMetadata, error) {
	format := fmt.Sprintf("%s_${container name}_${sandbox name}_${sandbox namespace}_${sandbox uid}_${attempt times}", kubePrefix)

	parts := strings.Split(name, nameDelimiter)
	if len(parts) != 6 {
		return nil, fmt.Errorf("failed to parse container name: %q, which should be %s", name, format)
	}
	if parts[0] != kubePrefix {
		return nil, fmt.Errorf("container is not managed by kubernetes: %q", name)
	}

	attempt, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("failed to parse the attempt times in container name: %q: %v", name, err)
	}

	return &runtime.ContainerMetadata{
		Name:    parts[1],
		Attempt: attempt,
	}, nil
}

// modifyContainerNamespaceOptions apply namespace options for container.
func modifyContainerNamespaceOptions(nsOpts *runtime.NamespaceOption, podSandboxID string, hostConfig *apitypes.HostConfig) {
	sandboxNSMode := fmt.Sprintf("container:%v", podSandboxID)

	hostConfig.PidMode = sandboxNSMode
	hostConfig.NetworkMode = sandboxNSMode
	hostConfig.IpcMode = sandboxNSMode
	hostConfig.UTSMode = sandboxNSMode
}

// applyContainerSecurityContext updates pouch container options according to security context.
func applyContainerSecurityContext(lc *runtime.LinuxContainerConfig, podSandboxID string, config *apitypes.ContainerConfig, hc *apitypes.HostConfig) error {
	// TODO: modify Config and HostConfig.

	modifyContainerNamespaceOptions(lc.SecurityContext.GetNamespaceOptions(), podSandboxID, hc)

	return nil
}

// Apply Linux-specific options if applicable.
func (c *CriManager) updateCreateConfig(createConfig *apitypes.ContainerCreateConfig, config *runtime.ContainerConfig, sandboxConfig *runtime.PodSandboxConfig, podSandboxID string) error {
	if lc := config.GetLinux(); lc != nil {
		// TODO: resource restriction.

		// Apply security context.
		if err := applyContainerSecurityContext(lc, podSandboxID, &createConfig.ContainerConfig, createConfig.HostConfig); err != nil {
			return fmt.Errorf("failed to apply container security context for container %q: %v", config.Metadata.Name, err)
		}
	}

	// TODO: apply cgroupParent derived from the sandbox config.

	return nil
}

func toCriContainerState(status apitypes.Status) runtime.ContainerState {
	switch status {
	case apitypes.StatusRunning:
		return runtime.ContainerState_CONTAINER_RUNNING
	case apitypes.StatusExited:
		return runtime.ContainerState_CONTAINER_EXITED
	case apitypes.StatusCreated:
		return runtime.ContainerState_CONTAINER_CREATED
	default:
		return runtime.ContainerState_CONTAINER_UNKNOWN
	}
}

func toCriContainer(c *ContainerMeta) (*runtime.Container, error) {
	state := toCriContainerState(c.State.Status)
	metadata, err := parseContainerName(c.Name)
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Config.Labels)
	sandboxID := c.Config.Labels[sandboxIDLabelKey]

	return &runtime.Container{
		Id:           c.ID,
		PodSandboxId: sandboxID,
		Metadata:     metadata,
		Image:        &runtime.ImageSpec{Image: c.Config.Image},
		ImageRef:     c.Image,
		State:        state,
		// TODO: fill "CreatedAt" when it is appropriate.
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

func filterCRIContainers(containers []*runtime.Container, filter *runtime.ContainerFilter) []*runtime.Container {
	if filter == nil {
		return containers
	}

	filtered := []*runtime.Container{}
	for _, c := range containers {
		if filter.GetId() != "" && filter.GetId() != c.Id {
			continue
		}
		if filter.GetPodSandboxId() != "" && filter.GetPodSandboxId() != c.PodSandboxId {
			continue
		}
		if filter.GetState() != nil && filter.GetState().GetState() != c.State {
			continue
		}
		if filter.GetLabelSelector() != nil {
			match := true
			for k, v := range filter.GetLabelSelector() {
				value, ok := c.Labels[k]
				if !ok || v != value {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, c)
	}

	return filtered
}

// Image related tool functions.

// imageToCriImage converts pouch image API to CRI image API.
func imageToCriImage(image *apitypes.ImageInfo) (*runtime.Image, error) {
	ref, err := reference.Parse(image.Name)
	if err != nil {
		return nil, err
	}

	uid := &runtime.Int64Value{}
	imageUID, username := getUserFromImageUser(image.Config.User)
	if imageUID != nil {
		uid.Value = *imageUID
	}

	size := uint64(image.Size)
	// TODO: improve type ImageInfo to include RepoTags and RepoDigests.
	return &runtime.Image{
		Id:          image.Digest,
		RepoTags:    []string{fmt.Sprintf("%s:%s", ref.Name, ref.Tag)},
		RepoDigests: []string{fmt.Sprintf("%s@%s", ref.Name, image.Digest)},
		Size_:       size,
		Uid:         uid,
		Username:    username,
	}, nil
}

// ensureSandboxImageExists pulls the image when it's not present.
func (c *CriManager) ensureSandboxImageExists(ctx context.Context, image string) error {
	_, err := c.ImageMgr.GetImage(ctx, image)
	// TODO: maybe we should distinguish NotFound error with others.
	if err == nil {
		return nil
	}

	ref, err := reference.Parse(image)
	if err != nil {
		return fmt.Errorf("parse image name failed: %v", err)
	}

	err = c.ImageMgr.PullImage(ctx, ref.Name, ref.Tag, bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("pull sandbox image %q failed: %v", image, err)
	}

	return nil
}

// getUserFromImageUser gets uid or user name of the image user.
// If user is numeric, it will be treated as uid; or else, it is treated as user name.
func getUserFromImageUser(imageUser string) (*int64, string) {
	user := parseUserFromImageUser(imageUser)
	// return both nil if user is not specified in the image.
	if user == "" {
		return nil, ""
	}
	// user could be either uid or user name. Try to interpret as numeric uid.
	uid, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		// If user is non numeric, assume it's user name.
		return nil, user
	}
	// If user is a numeric uid.
	return &uid, ""
}

// parseUserFromImageUser splits the user out of an user:group string.
func parseUserFromImageUser(id string) string {
	if id == "" {
		return id
	}
	// split instances where the id may contain user:group
	if strings.Contains(id, ":") {
		return strings.Split(id, ":")[0]
	}
	// no group, just return the id
	return id
}
