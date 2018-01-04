package mgr

import (
	"context"
	"fmt"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// isEmpty indicates whether namespace mode is empty.
func isEmpty(mode string) bool {
	return mode == ""
}

// isNone indicates whether container's namespace mode is set to "none".
func isNone(mode string) bool {
	return mode == "none"
}

// isHost indicates whether the container shares the host's corresponding namespace.
func isHost(mode string) bool {
	return mode == "host"
}

// isShareable indicates whether the containers namespace can be shared with another container.
func isShareable(mode string) bool {
	return mode == "shareable"
}

// isContainer indicates whether the container uses another container's corresponding namespace.
func isContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// isPrivate indicates whether the container uses its own namespace.
func isPrivate(ns specs.LinuxNamespaceType, mode string) bool {
	switch ns {
	case specs.IPCNamespace:
		return mode == "private"
	case specs.NetworkNamespace, specs.PIDNamespace:
		return !(isHost(mode) || isContainer(mode))
	case specs.UserNamespace, specs.UTSNamespace:
		return !(isHost(mode))
	}
	return false
}

// connectedContainer is the id or name of the container whose namespace this container share with.
func connectedContainer(mode string) string {
	parts := strings.SplitN(mode, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func getIpcContainer(ctx context.Context, mgr ContainerMgr, id string) (*ContainerMeta, error) {
	// Check whether the container exists.
	c, err := mgr.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't join IPC namespace of container %q: %v", id, err)
	}

	// TODO: check whether the container is running and not restarting.

	// TODO: check whether the container's ipc namespace is shareable.

	return c, nil
}

func getPidContainer(ctx context.Context, mgr ContainerMgr, id string) (*ContainerMeta, error) {
	// Check the container exists.
	c, err := mgr.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't join PID namespace of %q: %v", id, err)
	}

	// TODO: check whether the container is running and not restarting.

	return c, nil
}

// TODO
func setupUserNamespace(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	return nil
}

// TODO
func setupNetworkNamespace(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	return nil
}

func setupIpcNamespace(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	ipcMode := meta.HostConfig.IpcMode
	switch {
	case isContainer(ipcMode):
		ns := specs.LinuxNamespace{Type: specs.IPCNamespace}
		c, err := getIpcContainer(ctx, spec.ctrMgr, connectedContainer(ipcMode))
		if err != nil {
			return fmt.Errorf("setup container ipc namespace mode failed: %v", err)
		}
		ns.Path = fmt.Sprintf("/proc/%d/ns/ipc", c.State.Pid)
		setNamespace(spec.s, ns)
	case isHost(ipcMode):
		removeNamespace(spec.s, specs.IPCNamespace)
	default:
		ns := specs.LinuxNamespace{Type: specs.IPCNamespace}
		setNamespace(spec.s, ns)
	}
	return nil
}

func setupPidNamespace(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	pidMode := meta.HostConfig.PidMode
	switch {
	case isContainer(pidMode):
		ns := specs.LinuxNamespace{Type: specs.PIDNamespace}
		c, err := getPidContainer(ctx, spec.ctrMgr, connectedContainer(pidMode))
		if err != nil {
			return fmt.Errorf("setup container pid namespace mode failed: %v", err)
		}
		ns.Path = fmt.Sprintf("/proc/%d/ns/pid", c.State.Pid)
		setNamespace(spec.s, ns)
	case isHost(pidMode):
		removeNamespace(spec.s, specs.PIDNamespace)
	default:
		ns := specs.LinuxNamespace{Type: specs.PIDNamespace}
		setNamespace(spec.s, ns)
	}
	return nil
}

func setupUtsNamespace(ctx context.Context, meta *ContainerMeta, spec *SpecWrapper) error {
	utsMode := meta.HostConfig.UTSMode
	switch {
	case isHost(utsMode):
		removeNamespace(spec.s, specs.UTSNamespace)
	default:
		ns := specs.LinuxNamespace{Type: specs.UTSNamespace}
		setNamespace(spec.s, ns)
		// set hostname
		if hostname := meta.Config.Hostname.String(); hostname != "" {
			spec.s.Hostname = hostname
		}
	}
	return nil
}

func setNamespace(s *specs.Spec, ns specs.LinuxNamespace) {
	for i, n := range s.Linux.Namespaces {
		if n.Type == ns.Type {
			s.Linux.Namespaces[i] = ns
			return
		}
	}
	s.Linux.Namespaces = append(s.Linux.Namespaces, ns)
}

func removeNamespace(s *specs.Spec, nsType specs.LinuxNamespaceType) {
	for i, n := range s.Linux.Namespaces {
		if n.Type == nsType {
			s.Linux.Namespaces = append(s.Linux.Namespaces[:i], s.Linux.Namespaces[i+1:]...)
			return
		}
	}
}
