package stream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/daemon/mgr"
	pkgstreams "github.com/alibaba/pouch/pkg/streams"

	"github.com/sirupsen/logrus"
)

// Runtime is the interface to execute the commands and provide the streams.
type Runtime interface {
	// Exec executes the command in pod.
	Exec(ctx context.Context, containerID string, cmd []string, resizeChan <-chan apitypes.ResizeOptions, streamOpts *remotecommand.Options, streams *remotecommand.Streams) (uint32, error)

	// Attach attaches to pod.
	Attach(ctx context.Context, containerID string, streamOpts *remotecommand.Options, streams *remotecommand.Streams) error

	// PortForward forward port to pod.
	PortForward(ctx context.Context, name string, port int32, stream io.ReadWriteCloser) error
}

type streamRuntime struct {
	containerMgr mgr.ContainerMgr
}

// NewStreamRuntime creates a brand new stream runtime.
func NewStreamRuntime(ctrMgr mgr.ContainerMgr) Runtime {
	return &streamRuntime{containerMgr: ctrMgr}
}

// Exec executes a command inside the container.
func (s *streamRuntime) Exec(ctx context.Context, containerID string, cmd []string, resizeChan <-chan apitypes.ResizeOptions, streamOpts *remotecommand.Options, streams *remotecommand.Streams) (uint32, error) {
	createConfig := &apitypes.ExecCreateConfig{
		Cmd:          cmd,
		AttachStdin:  streamOpts.Stdin,
		AttachStdout: streamOpts.Stdout,
		AttachStderr: streamOpts.Stderr,
		Tty:          streamOpts.TTY,
	}

	execid, err := s.containerMgr.CreateExec(ctx, containerID, createConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to create exec for container %q: %v", containerID, err)
	}

	handleResizing(containerID, execid, resizeChan, func(size apitypes.ResizeOptions) {
		err := s.containerMgr.ResizeExec(ctx, execid, size)
		if err != nil {
			logrus.Errorf("failed to resize process %q console for container %q: %v", execid, containerID, err)
		}
	})

	attachCfg := &pkgstreams.AttachConfig{
		UseStdin:  createConfig.AttachStdin,
		Stdin:     streams.StdinStream,
		UseStdout: createConfig.AttachStdout,
		Stdout:    streams.StdoutStream,
		UseStderr: createConfig.AttachStderr,
		Stderr:    streams.StderrStream,
		Terminal:  createConfig.Tty,
	}

	if err := s.containerMgr.StartExec(ctx, execid, attachCfg); err != nil {
		return 0, fmt.Errorf("failed to exec for container %q: %v", containerID, err)
	}

	ei, err := s.containerMgr.InspectExec(ctx, execid)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect exec for container %q: %v", containerID, err)
	}
	return uint32(ei.ExitCode), nil
}

// handleResizing spawns a goroutine that processes the resize channel, calling resizeFunc for each
// remotecommand.TerminalSize received from the channel. The resize channel must be closed elsewhere to stop the
// goroutine.
func handleResizing(containerID, execID string, resizeChan <-chan apitypes.ResizeOptions, resizeFunc func(size apitypes.ResizeOptions)) {
	if resizeChan == nil {
		return
	}
	go func() {
		for {
			size, ok := <-resizeChan
			if !ok {
				return
			}
			if size.Height <= 0 || size.Width <= 0 {
				continue
			}
			resizeFunc(size)
		}
	}()
}

// Attach attaches to a running container.
func (s *streamRuntime) Attach(ctx context.Context, containerID string, streamOpts *remotecommand.Options, streams *remotecommand.Streams) error {
	// TODO(fuweid): could we close stdin after stop attach?
	attachCfg := &pkgstreams.AttachConfig{
		UseStdin:  streamOpts.Stdin,
		Stdin:     streams.StdinStream,
		UseStdout: streamOpts.Stdout,
		Stdout:    streams.StdoutStream,
		UseStderr: streamOpts.Stderr,
		Stderr:    streams.StderrStream,
		Terminal:  streamOpts.TTY,
	}
	if err := s.containerMgr.AttachContainerIO(ctx, containerID, attachCfg); err != nil {
		return fmt.Errorf("failed to attach to container %q: %v", containerID, err)
	}
	return nil
}

// PortForward forwards ports from a PodSandbox.
func (s *streamRuntime) PortForward(ctx context.Context, id string, port int32, stream io.ReadWriteCloser) error {
	sandbox, err := s.containerMgr.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get metadata of sandbox %q: %v", id, err)
	}
	pid := sandbox.State.Pid

	socat, err := exec.LookPath("socat")
	if err != nil {
		return fmt.Errorf("failed to find socat: %v", err)
	}

	// Check following links for meaning of the options:
	// * socat: https://linux.die.net/man/1/socat
	// * nsenter: http://man7.org/linux/man-pages/man1/nsenter.1.html
	args := []string{"-t", fmt.Sprintf("%d", pid), "-n", socat,
		"-", fmt.Sprintf("TCP4:localhost:%d", port)}

	nsenter, err := exec.LookPath("nsenter")
	if err != nil {
		return fmt.Errorf("failed to find nsenter: %v", err)
	}

	logrus.Infof("Executing port forwarding command: %s %s", nsenter, strings.Join(args, " "))

	cmd := exec.Command(nsenter, args...)
	cmd.Stdout = stream

	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	// If we use Stdin, command.Run() won't return until the goroutine that's copying
	// from stream finishes. Unfortunately, if you have a client like telnet connected
	// via port forwarding, as long as the user's telnet client is connected to the user's
	// local listener that port forwarding sets up, the telnet session never exits. This
	// means that even if socat has finished running, command.Run() won't ever return
	// (because the client still has the connection and stream open).
	//
	// The work around is to use StdinPipe(), as Wait() (called by Run()) closes the pipe
	// when the command (socat) exits.
	in, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	go func() {
		if _, err := io.Copy(in, stream); err != nil {
			logrus.Errorf("failed to copy port forward input for %q port %d: %v", id, port, err)
		}
		in.Close()
		logrus.Infof("finish copy port forward input for %q port %d: %v", id, port, err)
	}()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("nsenter command returns error: %v, stderr: %q", err, stderr.String())
	}

	logrus.Infof("finish port forwarding for %q port %d", id, port)

	return nil
}
