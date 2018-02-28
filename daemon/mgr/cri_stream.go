package mgr

import (
	"net"
	"fmt"

	"github.com/alibaba/pouch/cri/stream"
)

func newStreamServer(address string, port string) (stream.Server, error){
	config := stream.DefaultConfig
	config.Address = net.JoinHostPort(address, port)
	runtime := newStreamRuntime()
	return stream.NewServer(config, runtime)
}

type streamRuntime struct {}

func newStreamRuntime() stream.Runtime {
	return &streamRuntime{}
}

// Exec executes a command inside the container.
func (s *streamRuntime) Exec() error {
	return fmt.Errorf("streamRuntime's Exec Not Implemented Yet")
}

// Attach attaches to a running container.
func (s *streamRuntime) Attach() error {
	return fmt.Errorf("streamRuntime's Attach Not Implemented Yet")
}

// PortForward forwards ports from a PodSandbox.
func (s *streamRuntime) PortForward() error {
	return fmt.Errorf("streamRuntime's PortForward Not Implemented Yet")
}
