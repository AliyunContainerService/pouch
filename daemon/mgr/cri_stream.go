package mgr

import (
	"fmt"
	"net"
	"io"
	"net/http"

	"github.com/alibaba/pouch/cri/stream"
)

func newStreamServer(address string, port string) (stream.Server, error) {
	config := stream.DefaultConfig
	config.Address = net.JoinHostPort(address, port)
	runtime := newStreamRuntime()
	return stream.NewServer(config, runtime)
}

type streamRuntime struct{}

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
func (s *streamRuntime) PortForward(name string, port int32, stream io.ReadWriteCloser) error {
	res, err := http.Get("http://www.baidu.com")
	if err != nil {
		return fmt.Errorf("failed to http get baidu: %v", err)
	}

	res.Write(stream)

	return nil
}
