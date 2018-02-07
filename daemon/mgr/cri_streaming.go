package mgr

import (
	"fmt"
	"io"
	"net"

	k8snet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/pkg/kubelet/server/streaming"
)

func newStreamServer(ctrMgr ContainerMgr, address, port string) (streaming.Server, error) {
	if address == "" {
		// If address is "", we would use the host's default interface.
		addr, err := k8snet.ChooseBindAddress(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get stream server address: %v", err)
		}
		address = addr.String()
	}

	config := streaming.DefaultConfig
	config.Addr = net.JoinHostPort(address, port)
	runtime := newStreamRuntime(ctrMgr)
	return streaming.NewServer(config, runtime)
}

type streamRuntime struct {
	containerMgr ContainerMgr
}

func newStreamRuntime(ctrMgr ContainerMgr) streaming.Runtime {
	return &streamRuntime{containerMgr: ctrMgr}
}

// Exec executes a command inside the container.
func (s *streamRuntime) Exec(containerID string, cmd []string, stdin io.Reader, stdout, stderr io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error {
	return fmt.Errorf("streamRuntime's Exec Not Implemented Yet")
}

// Attach attaches to a running container.
func (s *streamRuntime) Attach(containerID string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error {
	return fmt.Errorf("streamRuntime's Attach Not Implemented Yet")
}

// PortForward forwards ports from a PodSandbox.
func (s *streamRuntime) PortForward(podSandboxID string, port int32, stream io.ReadWriteCloser) error {
	return fmt.Errorf("streamRuntime's PortForward Not Implemented Yet")
}
