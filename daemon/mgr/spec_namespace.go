package mgr

import (
	"context"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// TODO
func setupUserNamespace(ctx context.Context, meta *ContainerMeta, s *specs.Spec) error {
	return nil
}

// TODO
func setupNetworkNamespace(ctx context.Context, meta *ContainerMeta, s *specs.Spec) error {
	return nil
}

func setupIpcNamespace(ctx context.Context, meta *ContainerMeta, s *specs.Spec) error {
	ns := specs.LinuxNamespace{Type: specs.IPCNamespace}
	setNamespace(s, ns)
	return nil
}

func setupPidNamespace(ctx context.Context, meta *ContainerMeta, s *specs.Spec) error {
	ns := specs.LinuxNamespace{Type: specs.PIDNamespace}
	setNamespace(s, ns)
	return nil
}

func setupUtsNamespace(ctx context.Context, meta *ContainerMeta, s *specs.Spec) error {
	ns := specs.LinuxNamespace{Type: specs.UTSNamespace}
	setNamespace(s, ns)
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
