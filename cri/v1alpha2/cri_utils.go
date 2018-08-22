package v1alpha2

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
	anno "github.com/alibaba/pouch/cri/annotations"
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/cgroups"
	"github.com/containerd/typeurl"
	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/go-openapi/strfmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
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
	return result
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

// generateContainerMounts sets up necessary container mounts including /dev/shm, /etc/hosts
// and /etc/resolv.conf.
func generateContainerMounts(sandboxRootDir string) []string {
	// TODO: more attr and check whether these bindings is included in cri mounts.
	result := []string{}
	hostPath := path.Join(sandboxRootDir, "resolv.conf")
	containerPath := resolvConfPath
	result = append(result, fmt.Sprintf("%s:%s", hostPath, containerPath))

	return result
}

func generateMountBindings(mounts []*runtime.Mount) []string {
	result := make([]string, 0, len(mounts))
	for _, m := range mounts {
		bind := fmt.Sprintf("%s:%s", m.HostPath, m.ContainerPath)
		var attrs []string
		if m.Readonly {
			attrs = append(attrs, "ro")
		}
		if m.SelinuxRelabel {
			attrs = append(attrs, "Z")
		}
		switch m.Propagation {
		case runtime.MountPropagation_PROPAGATION_PRIVATE:
			// noop, default mode is private.
		case runtime.MountPropagation_PROPAGATION_BIDIRECTIONAL:
			attrs = append(attrs, "rshared")
		case runtime.MountPropagation_PROPAGATION_HOST_TO_CONTAINER:
			attrs = append(attrs, "rslave")
		}
		if len(attrs) > 0 {
			bind = fmt.Sprintf("%s:%s", bind, strings.Join(attrs, ","))
		}
		result = append(result, bind)
	}
	return result
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

func modifySandboxNamespaceOptions(nsOpts *runtime.NamespaceOption, hostConfig *apitypes.HostConfig) {
	if nsOpts == nil {
		return
	}
	if nsOpts.GetPid() == runtime.NamespaceMode_NODE {
		hostConfig.PidMode = namespaceModeHost
	}
	if nsOpts.GetIpc() == runtime.NamespaceMode_NODE {
		hostConfig.IpcMode = namespaceModeHost
	}
	if nsOpts.GetNetwork() == runtime.NamespaceMode_NODE {
		hostConfig.NetworkMode = namespaceModeHost
	}
}

func applySandboxSecurityContext(lc *runtime.LinuxPodSandboxConfig, config *apitypes.ContainerConfig, hc *apitypes.HostConfig) error {
	if lc == nil {
		return nil
	}

	var sc *runtime.LinuxContainerSecurityContext
	if lc.SecurityContext != nil {
		sc = &runtime.LinuxContainerSecurityContext{
			SupplementalGroups: lc.SecurityContext.SupplementalGroups,
			RunAsUser:          lc.SecurityContext.RunAsUser,
			ReadonlyRootfs:     lc.SecurityContext.ReadonlyRootfs,
			SelinuxOptions:     lc.SecurityContext.SelinuxOptions,
			NamespaceOptions:   lc.SecurityContext.NamespaceOptions,
		}
	}

	modifyContainerConfig(sc, config)
	err := modifyHostConfig(sc, hc)
	if err != nil {
		return err
	}
	modifySandboxNamespaceOptions(sc.GetNamespaceOptions(), hc)

	return nil
}

// applySandboxLinuxOptions applies LinuxPodSandboxConfig to pouch's HostConfig and ContainerCreateConfig.
func applySandboxLinuxOptions(hc *apitypes.HostConfig, lc *runtime.LinuxPodSandboxConfig, createConfig *apitypes.ContainerCreateConfig, image string) error {
	// apply the sandbox network_mode, "none" is default.
	hc.NetworkMode = namespaceModeNone

	if lc == nil {
		return nil
	}

	// Apply security context.
	err := applySandboxSecurityContext(lc, &createConfig.ContainerConfig, hc)
	if err != nil {
		return err
	}

	// Set sysctls.
	hc.Sysctls = lc.Sysctls
	return nil
}

// makeSandboxPouchConfig returns apitypes.ContainerCreateConfig based on runtime.PodSandboxConfig.
func makeSandboxPouchConfig(config *runtime.PodSandboxConfig, image string) (*apitypes.ContainerCreateConfig, error) {
	// Merge annotations and labels because pouch supports only labels.
	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply a label to distinguish sandboxes from regular containers.
	labels[containerTypeLabelKey] = containerTypeLabelSandbox

	hc := &apitypes.HostConfig{}

	// Apply runtime options.
	if annotations := config.GetAnnotations(); annotations != nil {
		hc.Runtime = annotations[anno.KubernetesRuntime]
	}

	createConfig := &apitypes.ContainerCreateConfig{
		ContainerConfig: apitypes.ContainerConfig{
			Hostname: strfmt.Hostname(config.Hostname),
			Image:    image,
			Labels:   labels,
		},
		HostConfig:       hc,
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}

	// Apply linux-specific options.
	err := applySandboxLinuxOptions(hc, config.GetLinux(), createConfig, image)
	if err != nil {
		return nil, err
	}
	// Apply resource options.
	if lc := config.GetLinux(); lc != nil {
		hc.CgroupParent = lc.CgroupParent
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

func toCriSandbox(c *mgr.Container) (*runtime.PodSandbox, error) {
	state := toCriSandboxState(c.State.Status)
	metadata, err := parseSandboxName(c.Name)
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Config.Labels)

	createdAt, err := toCriTimestamp(c.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to parse create timestamp for container %q: %v", c.ID, err)
	}

	return &runtime.PodSandbox{
		Id:          c.ID,
		Metadata:    metadata,
		State:       state,
		CreatedAt:   createdAt,
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

// It has the possibility that we failed to run the sandbox and it is not being cleaned up.
// Kubelet will use list to get the sandboxes, but will not get the status of the failed pod
// whose meta data has not been put into the Sandbox Store. And Kubelet will keep trying to
// get the status of the failed pod and won't create a new one to replace it. It's a DEAD LOCK.
// Actually Kubelet should not know the existence of invalid pod whose meta data won't be in the
// Sandbox Store. So we could avoid the DEAD LOCK mentioned above.
func (c *CriManager) filterInvalidSandboxes(ctx context.Context, sandboxes []*mgr.Container) ([]*mgr.Container, error) {
	validSandboxes, err := c.SandboxStore.Keys()
	if err != nil {
		return nil, err
	}

	var result []*mgr.Container
	for _, sandbox := range sandboxes {
		exist := false
		for _, id := range validSandboxes {
			if sandbox.ID == id {
				exist = true
				break
			}
		}
		if exist {
			result = append(result, sandbox)
			continue
		}

		status := sandbox.State.Status
		// NOTE: what if the worst case that we failed to remove the sandbox and
		// it is still running?
		if status != apitypes.StatusRunning && status != apitypes.StatusCreated {
			logrus.Warnf("filterInvalidSandboxes: remove invalid sandbox %v", sandbox.ID)
			c.ContainerMgr.Remove(ctx, sandbox.ID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
		}
	}
	return result, nil
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

// parseDNSOptions parse DNS options into resolv.conf format content,
// if none option is specified, will return empty with no error.
func parseDNSOptions(servers, searches, options []string) (string, error) {
	resolvContent := ""

	if len(searches) > 0 {
		resolvContent += fmt.Sprintf("search %s\n", strings.Join(searches, " "))
	}

	if len(servers) > 0 {
		resolvContent += fmt.Sprintf("nameserver %s\n", strings.Join(servers, "\nnameserver "))
	}

	if len(options) > 0 {
		resolvContent += fmt.Sprintf("options %s\n", strings.Join(options, " "))
	}

	return resolvContent, nil
}

// copyFile copys src file to dest file
func copyFile(src, dest string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// setupSandboxFiles sets up necessary sandbox files.
func setupSandboxFiles(sandboxRootDir string, config *runtime.PodSandboxConfig) error {
	// Set DNS options. Maintain a resolv.conf for the sandbox.
	var resolvContent string
	resolvPath := path.Join(sandboxRootDir, "resolv.conf")

	var err error
	dnsConfig := config.GetDnsConfig()
	if dnsConfig != nil {
		resolvContent, err = parseDNSOptions(dnsConfig.Servers, dnsConfig.Searches, dnsConfig.Options)
		if err != nil {
			return fmt.Errorf("failed to parse sandbox DNSConfig %+v: %v", dnsConfig, err)
		}
	}

	if resolvContent == "" {
		// Copy host's resolv.conf to resolvPath.
		err = copyFile(resolvConfPath, resolvPath, 0644)
		if err != nil {
			return fmt.Errorf("failed to copy host's resolv.conf to %q: %v", resolvPath, err)
		}
	} else {
		err = ioutil.WriteFile(resolvPath, []byte(resolvContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write resolv content to %q: %v", resolvPath, err)
		}
	}

	return nil
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

// makeupLogPath makes up the log path of container from log directory and its metadata.
func makeupLogPath(logDirectory string, metadata *runtime.ContainerMetadata) string {
	return filepath.Join(logDirectory, metadata.Name, fmt.Sprintf("%d.log", metadata.Attempt))
}

// modifyContainerNamespaceOptions apply namespace options for container.
func modifyContainerNamespaceOptions(nsOpts *runtime.NamespaceOption, podSandboxID string, hostConfig *apitypes.HostConfig) {
	sandboxNSMode := fmt.Sprintf("container:%v", podSandboxID)

	if nsOpts == nil {
		hostConfig.PidMode = sandboxNSMode
		hostConfig.IpcMode = sandboxNSMode
		hostConfig.NetworkMode = sandboxNSMode
		return
	}

	for _, n := range []struct {
		hostMode bool
		nsMode   *string
	}{
		{
			hostMode: nsOpts.GetPid() == runtime.NamespaceMode_NODE,
			nsMode:   &hostConfig.PidMode,
		},
		{
			hostMode: nsOpts.GetIpc() == runtime.NamespaceMode_NODE,
			nsMode:   &hostConfig.IpcMode,
		},
		{
			hostMode: nsOpts.GetNetwork() == runtime.NamespaceMode_NODE,
			nsMode:   &hostConfig.NetworkMode,
		},
	} {
		if n.hostMode {
			*n.nsMode = namespaceModeHost
		} else {
			if n.nsMode == &hostConfig.PidMode && nsOpts.GetPid() == runtime.NamespaceMode_CONTAINER {
				continue
			}
			*n.nsMode = sandboxNSMode
		}
	}
}

// getAppArmorSecurityOpts gets appArmor options from container config.
func getAppArmorSecurityOpts(sc *runtime.LinuxContainerSecurityContext) ([]string, error) {
	profile := sc.ApparmorProfile
	if profile == "" || profile == mgr.ProfileRuntimeDefault {
		// Pouch should applies the default profile by default.
		return nil, nil
	}

	// Return unconfined profile explicitly.
	if profile == mgr.ProfileNameUnconfined {
		return []string{fmt.Sprintf("apparmor=%s", profile)}, nil
	}

	if !strings.HasPrefix(profile, mgr.ProfileNamePrefix) {
		return nil, fmt.Errorf("undefault profile name should prefix with %q", mgr.ProfileNamePrefix)
	}
	profile = strings.TrimPrefix(profile, mgr.ProfileNamePrefix)

	return []string{fmt.Sprintf("apparmor=%s", profile)}, nil
}

func getSELinuxSecurityOpts(sc *runtime.LinuxContainerSecurityContext) ([]string, error) {
	if sc.SelinuxOptions == nil {
		return nil, nil
	}

	var result []string
	selinuxOpts := sc.SelinuxOptions
	// Should ignore incomplete selinuxOpts.
	if selinuxOpts.GetUser() == "" ||
		selinuxOpts.GetRole() == "" ||
		selinuxOpts.GetType() == "" ||
		selinuxOpts.GetLevel() == "" {
		return nil, nil
	}

	for k, v := range map[string]string{
		"user":  selinuxOpts.User,
		"role":  selinuxOpts.Role,
		"type":  selinuxOpts.Type,
		"level": selinuxOpts.Level,
	} {
		if len(v) > 0 {
			result = append(result, fmt.Sprintf("label=%s:%s", k, v))
		}
	}

	return result, nil
}

// getSeccompSecurityOpts get container seccomp options from container seccomp profiles.
func getSeccompSecurityOpts(sc *runtime.LinuxContainerSecurityContext) ([]string, error) {
	profile := sc.SeccompProfilePath
	if profile == "" || profile == mgr.ProfileNameUnconfined {
		return []string{fmt.Sprintf("seccomp=%s", mgr.ProfileNameUnconfined)}, nil
	}

	// Return unconfined profile explicitly.
	if profile == mgr.ProfileDockerDefault {
		// return nil so pouch will load the default seccomp profile.
		return nil, nil
	}

	if !strings.HasPrefix(profile, mgr.ProfileNamePrefix) {
		return nil, fmt.Errorf("undefault profile %q should prefix with %q", profile, mgr.ProfileNamePrefix)
	}
	profile = strings.TrimPrefix(profile, mgr.ProfileNamePrefix)

	return []string{fmt.Sprintf("seccomp=%s", profile)}, nil
}

// modifyHostConfig applies security context config to pouch's HostConfig.
func modifyHostConfig(sc *runtime.LinuxContainerSecurityContext, hostConfig *apitypes.HostConfig) error {
	if sc == nil {
		return nil
	}

	// Apply supplemental groups.
	for _, group := range sc.SupplementalGroups {
		hostConfig.GroupAdd = append(hostConfig.GroupAdd, strconv.FormatInt(group, 10))
	}

	// TODO: apply other security options.

	// Apply capability options.
	hostConfig.Privileged = sc.Privileged
	hostConfig.ReadonlyRootfs = sc.ReadonlyRootfs
	if sc.GetCapabilities() != nil {
		hostConfig.CapAdd = sc.GetCapabilities().GetAddCapabilities()
		hostConfig.CapDrop = sc.GetCapabilities().GetDropCapabilities()
	}

	// Apply seccomp options.
	seccompSecurityOpts, err := getSeccompSecurityOpts(sc)
	if err != nil {
		return fmt.Errorf("failed to generate seccomp security options: %v", err)
	}
	hostConfig.SecurityOpt = append(hostConfig.SecurityOpt, seccompSecurityOpts...)

	// Apply SELinux options.
	selinuxSecurityOpts, err := getSELinuxSecurityOpts(sc)
	if err != nil {
		return fmt.Errorf("failed to generate SELinux security options: %v", err)
	}
	hostConfig.SecurityOpt = append(hostConfig.SecurityOpt, selinuxSecurityOpts...)

	// Apply appArmor options.
	appArmorSecurityOpts, err := getAppArmorSecurityOpts(sc)
	if err != nil {
		return fmt.Errorf("failed to generate appArmor security options: %v", err)
	}
	hostConfig.SecurityOpt = append(hostConfig.SecurityOpt, appArmorSecurityOpts...)

	if sc.NoNewPrivs {
		hostConfig.SecurityOpt = append(hostConfig.SecurityOpt, "no-new-privileges")
	}
	return nil
}

// modifyContainerConfig applies container security context config to pouch's Config.
func modifyContainerConfig(sc *runtime.LinuxContainerSecurityContext, config *apitypes.ContainerConfig) {
	if sc == nil {
		return
	}
	if sc.RunAsUser != nil {
		config.User = strconv.FormatInt(sc.GetRunAsUser().Value, 10)
	}
	if sc.RunAsUsername != "" {
		config.User = sc.RunAsUsername
	}
}

// applyContainerSecurityContext updates pouch container options according to security context.
func applyContainerSecurityContext(lc *runtime.LinuxContainerConfig, podSandboxID string, config *apitypes.ContainerConfig, hc *apitypes.HostConfig) error {
	modifyContainerConfig(lc.SecurityContext, config)

	err := modifyHostConfig(lc.SecurityContext, hc)
	if err != nil {
		return err
	}

	modifyContainerNamespaceOptions(lc.SecurityContext.GetNamespaceOptions(), podSandboxID, hc)

	return nil
}

// Apply Linux-specific options if applicable.
func (c *CriManager) updateCreateConfig(createConfig *apitypes.ContainerCreateConfig, config *runtime.ContainerConfig, sandboxConfig *runtime.PodSandboxConfig, podSandboxID string) error {
	// Apply runtime options.
	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)
	if sandboxMeta.Runtime != "" {
		createConfig.HostConfig.Runtime = sandboxMeta.Runtime
	}

	if lc := config.GetLinux(); lc != nil {
		resources := lc.GetResources()
		if resources != nil {
			createConfig.HostConfig.Resources.CPUPeriod = resources.GetCpuPeriod()
			createConfig.HostConfig.Resources.CPUQuota = resources.GetCpuQuota()
			createConfig.HostConfig.Resources.CPUShares = resources.GetCpuShares()
			createConfig.HostConfig.Resources.Memory = resources.GetMemoryLimitInBytes()
			createConfig.HostConfig.Resources.CpusetCpus = resources.GetCpusetCpus()
			createConfig.HostConfig.Resources.CpusetMems = resources.GetCpusetMems()
		}

		// Apply security context.
		if err := applyContainerSecurityContext(lc, podSandboxID, &createConfig.ContainerConfig, createConfig.HostConfig); err != nil {
			return fmt.Errorf("failed to apply container security context for container %q: %v", config.Metadata.Name, err)
		}
	}

	// Apply cgroupsParent derived from the sandbox config.
	if lc := sandboxConfig.GetLinux(); lc != nil {
		// Apply Cgroup options.
		createConfig.HostConfig.CgroupParent = lc.CgroupParent
	}

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

func toCriContainer(c *mgr.Container) (*runtime.Container, error) {
	state := toCriContainerState(c.State.Status)
	metadata, err := parseContainerName(c.Name)
	if err != nil {
		return nil, err
	}
	labels, annotations := extractLabels(c.Config.Labels)
	sandboxID := c.Config.Labels[sandboxIDLabelKey]

	createdAt, err := toCriTimestamp(c.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to parse create timestamp for container %q: %v", c.ID, err)
	}

	return &runtime.Container{
		Id:           c.ID,
		PodSandboxId: sandboxID,
		Metadata:     metadata,
		Image:        &runtime.ImageSpec{Image: c.Config.Image},
		ImageRef:     c.Image,
		State:        state,
		CreatedAt:    createdAt,
		Labels:       labels,
		Annotations:  annotations,
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

// containerNetns returns the network namespace of the given container.
func containerNetns(container *mgr.Container) string {
	pid := container.State.Pid
	if pid == -1 {
		// Pouch reports pid -1 for an exited container.
		return ""
	}

	return fmt.Sprintf("/proc/%v/ns/net", pid)
}

// Image related tool functions.

// imageToCriImage converts pouch image API to CRI image API.
func imageToCriImage(image *apitypes.ImageInfo) (*runtime.Image, error) {
	uid := &runtime.Int64Value{}
	imageUID, username := getUserFromImageUser(image.Config.User)
	if imageUID != nil {
		uid.Value = *imageUID
	}

	size := uint64(image.Size)
	// TODO: improve type ImageInfo to include RepoTags and RepoDigests.
	return &runtime.Image{
		Id:          image.ID,
		RepoTags:    image.RepoTags,
		RepoDigests: image.RepoDigests,
		Size_:       size,
		Uid:         uid,
		Username:    username,
		Volumes:     parseVolumesFromPouch(image.Config.Volumes),
	}, nil
}

// ensureSandboxImageExists pulls the image when it's not present.
func (c *CriManager) ensureSandboxImageExists(ctx context.Context, imageRef string) error {
	_, _, _, err := c.ImageMgr.CheckReference(ctx, imageRef)
	// TODO: maybe we should distinguish NotFound error with others.
	if err == nil {
		return nil
	}

	err = c.ImageMgr.PullImage(ctx, imageRef, nil, bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("pull sandbox image %q failed: %v", imageRef, err)
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

func (c *CriManager) attachLog(logPath string, containerID string, openStdin bool) error {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return fmt.Errorf("failed to create container for opening log file failed: %v", err)
	}
	// Attach to the container to get log.
	attachConfig := &mgr.AttachConfig{
		Stdin:      openStdin,
		Stdout:     true,
		Stderr:     true,
		CriLogFile: f,
	}
	err = c.ContainerMgr.Attach(context.Background(), containerID, attachConfig)
	if err != nil {
		return fmt.Errorf("failed to attach to container %q to get its log: %v", containerID, err)
	}
	return nil
}

func (c *CriManager) getContainerMetrics(ctx context.Context, meta *mgr.Container) (*runtime.ContainerStats, error) {
	var usedBytes, inodesUsed uint64

	stats, _, err := c.ContainerMgr.Stats(ctx, meta.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats of container %q: %v", meta.ID, err)
	}

	sn, err := c.SnapshotStore.Get(meta.ID)
	if err == nil {
		usedBytes = sn.Size
		inodesUsed = sn.Inodes
	}

	cs := &runtime.ContainerStats{}
	cs.WritableLayer = &runtime.FilesystemUsage{
		Timestamp: sn.Timestamp,
		FsId: &runtime.FilesystemIdentifier{
			Mountpoint: c.imageFSPath,
		},
		UsedBytes:  &runtime.UInt64Value{usedBytes},
		InodesUsed: &runtime.UInt64Value{inodesUsed},
	}

	metadata, err := parseContainerName(meta.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of container %q: %v", meta.ID, err)
	}

	labels, annotations := extractLabels(meta.Config.Labels)

	cs.Attributes = &runtime.ContainerAttributes{
		Id:          meta.ID,
		Metadata:    metadata,
		Labels:      labels,
		Annotations: annotations,
	}

	if stats != nil {
		s, err := typeurl.UnmarshalAny(stats.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract container metrics: %v", err)
		}
		metrics := s.(*cgroups.Metrics)
		if metrics.CPU != nil && metrics.CPU.Usage != nil {
			cs.Cpu = &runtime.CpuUsage{
				Timestamp:            stats.Timestamp.UnixNano(),
				UsageCoreNanoSeconds: &runtime.UInt64Value{metrics.CPU.Usage.Total},
			}
		}
		if metrics.Memory != nil && metrics.Memory.Usage != nil {
			cs.Memory = &runtime.MemoryUsage{
				Timestamp:       stats.Timestamp.UnixNano(),
				WorkingSetBytes: &runtime.UInt64Value{metrics.Memory.Usage.Usage},
			}
		}
	}

	return cs, nil
}

// imageFSPath returns containerd image filesystem path.
func imageFSPath(rootDir, snapshotter string) string {
	return filepath.Join(rootDir, fmt.Sprintf("%s.%s", snapshotPlugin, snapshotter))
}

// CRI extension related tool functions.

// parseResourceFromCRI parse Resources from runtime.LinuxContainerResources to apitypes.Resources
func parseResourcesFromCRI(runtimeResources *runtime.LinuxContainerResources) apitypes.Resources {
	var memorySwappiness *int64
	if runtimeResources.GetMemorySwappiness() != nil {
		memorySwappiness = &runtimeResources.GetMemorySwappiness().Value
	}

	return apitypes.Resources{
		CPUPeriod:            runtimeResources.GetCpuPeriod(),
		CPUQuota:             runtimeResources.GetCpuQuota(),
		CPUShares:            runtimeResources.GetCpuShares(),
		Memory:               runtimeResources.GetMemoryLimitInBytes(),
		CpusetCpus:           runtimeResources.GetCpusetCpus(),
		CpusetMems:           runtimeResources.GetCpusetMems(),
		BlkioWeight:          uint16(runtimeResources.GetBlkioWeight()),
		BlkioWeightDevice:    parseWeightDeviceFromCRI(runtimeResources.GetBlkioWeightDevice()),
		BlkioDeviceReadBps:   parseThrottleDeviceFromCRI(runtimeResources.GetBlkioDeviceReadBps()),
		BlkioDeviceWriteBps:  parseThrottleDeviceFromCRI(runtimeResources.GetBlkioDeviceWriteBps()),
		BlkioDeviceReadIOps:  parseThrottleDeviceFromCRI(runtimeResources.GetBlkioDeviceRead_IOps()),
		BlkioDeviceWriteIOps: parseThrottleDeviceFromCRI(runtimeResources.GetBlkioDeviceWrite_IOps()),
		KernelMemory:         runtimeResources.GetKernelMemory(),
		MemoryReservation:    runtimeResources.GetMemoryReservation(),
		MemorySwappiness:     memorySwappiness,
		Ulimits:              parseUlimitFromCRI(runtimeResources.GetUlimits()),
	}
}

// parseResourceFromPouch parse Resources from apitypes.Resources to runtime.LinuxContainerResources
func parseResourcesFromPouch(apitypesResources apitypes.Resources, diskQuota map[string]string) *runtime.LinuxContainerResources {
	var memorySwappiness *runtime.Int64Value
	if apitypesResources.MemorySwappiness != nil {
		memorySwappiness = &runtime.Int64Value{Value: *apitypesResources.MemorySwappiness}
	}

	return &runtime.LinuxContainerResources{
		CpuPeriod:             apitypesResources.CPUPeriod,
		CpuQuota:              apitypesResources.CPUQuota,
		CpuShares:             apitypesResources.CPUShares,
		MemoryLimitInBytes:    apitypesResources.Memory,
		CpusetCpus:            apitypesResources.CpusetCpus,
		CpusetMems:            apitypesResources.CpusetMems,
		BlkioWeight:           uint32(apitypesResources.BlkioWeight),
		BlkioWeightDevice:     parseWeightDeviceFromPouch(apitypesResources.BlkioWeightDevice),
		BlkioDeviceReadBps:    parseThrottleDeviceFromPouch(apitypesResources.BlkioDeviceReadBps),
		BlkioDeviceWriteBps:   parseThrottleDeviceFromPouch(apitypesResources.BlkioDeviceWriteBps),
		BlkioDeviceRead_IOps:  parseThrottleDeviceFromPouch(apitypesResources.BlkioDeviceReadIOps),
		BlkioDeviceWrite_IOps: parseThrottleDeviceFromPouch(apitypesResources.BlkioDeviceWriteIOps),
		KernelMemory:          apitypesResources.KernelMemory,
		MemoryReservation:     apitypesResources.MemoryReservation,
		MemorySwappiness:      memorySwappiness,
		Ulimits:               parseUlimitFromPouch(apitypesResources.Ulimits),
		DiskQuota:             diskQuota,
	}
}

// parseWeightDeviceFromCRI parse WeightDevice from runtime.WeightDevice to apitypes.WeightDevice
func parseWeightDeviceFromCRI(runtimeWeightDevices []*runtime.WeightDevice) (weightDevices []*apitypes.WeightDevice) {
	for _, v := range runtimeWeightDevices {
		weightDevices = append(weightDevices, &apitypes.WeightDevice{
			Path:   v.GetPath(),
			Weight: uint16(v.GetWeight()),
		})
	}
	return
}

// parseWeightDeviceFromPouch parse WeightDevice from apitypes.WeightDevice to runtime.WeightDevice
func parseWeightDeviceFromPouch(apitypesWeightDevices []*apitypes.WeightDevice) (weightDevices []*runtime.WeightDevice) {
	for _, v := range apitypesWeightDevices {
		weightDevices = append(weightDevices, &runtime.WeightDevice{
			Path:   v.Path,
			Weight: uint32(v.Weight),
		})
	}
	return
}

// parseThrottleDeviceFromCRI parse ThrottleDevice from runtime.ThrottleDevice to apitypes.ThrottleDevice
func parseThrottleDeviceFromCRI(runtimeThrottleDevices []*runtime.ThrottleDevice) (throttleDevices []*apitypes.ThrottleDevice) {
	for _, v := range runtimeThrottleDevices {
		throttleDevices = append(throttleDevices, &apitypes.ThrottleDevice{
			Path: v.GetPath(),
			Rate: v.GetRate(),
		})
	}
	return
}

// parseThrottleDeviceFromPouch parse ThrottleDevice from apitypes.ThrottleDevice to runtime.ThrottleDevice
func parseThrottleDeviceFromPouch(apitypesThrottleDevices []*apitypes.ThrottleDevice) (throttleDevices []*runtime.ThrottleDevice) {
	for _, v := range apitypesThrottleDevices {
		throttleDevices = append(throttleDevices, &runtime.ThrottleDevice{
			Path: v.Path,
			Rate: v.Rate,
		})
	}
	return
}

// parseUlimitFromCRI parse Ulimit from runtime.Ulimit to apitypes.Ulimit
func parseUlimitFromCRI(runtimeUlimits []*runtime.Ulimit) (ulimits []*apitypes.Ulimit) {
	for _, v := range runtimeUlimits {
		ulimits = append(ulimits, &apitypes.Ulimit{
			Hard: v.GetHard(),
			Name: v.GetName(),
			Soft: v.GetSoft(),
		})
	}
	return
}

// parseUlimitFromPouch parse Ulimit from apitypes.Ulimit to runtime.Ulimit
func parseUlimitFromPouch(apitypesUlimits []*apitypes.Ulimit) (ulimits []*runtime.Ulimit) {
	for _, v := range apitypesUlimits {
		ulimits = append(ulimits, &runtime.Ulimit{
			Hard: v.Hard,
			Name: v.Name,
			Soft: v.Soft,
		})
	}
	return
}

// parseVolumesFromPouch parse Volumes from map[string]interface{} to map[string]*runtime.Volume
func parseVolumesFromPouch(containerVolumes map[string]interface{}) map[string]*runtime.Volume {
	volumes := make(map[string]*runtime.Volume)
	for k := range containerVolumes {
		volumes[k] = &runtime.Volume{}
	}
	return volumes
}

// CNI Network related tool functions.

// toCNIPortMappings converts CRI port mappings to CNI.
func toCNIPortMappings(criPortMappings []*runtime.PortMapping) []ocicni.PortMapping {
	var portMappings []ocicni.PortMapping
	for _, mapping := range criPortMappings {
		if mapping.HostPort <= 0 {
			continue
		}
		portMappings = append(portMappings, ocicni.PortMapping{
			HostPort:      mapping.HostPort,
			ContainerPort: mapping.ContainerPort,
			Protocol:      strings.ToLower(mapping.Protocol.String()),
			HostIP:        mapping.HostIp,
		})
	}
	return portMappings
}
