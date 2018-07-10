package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/daemon/mgr"
	pouchnet "github.com/alibaba/pouch/pkg/net"

	"github.com/sirupsen/logrus"
)

func newStreamServer(ctrMgr mgr.ContainerMgr, address string, port string) (Server, error) {
	if address == "" {
		a, err := pouchnet.ChooseBindAddress(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get stream server address: %v", err)
		}
		address = a.String()
	}
	config := DefaultConfig
	config.Address = net.JoinHostPort(address, port)
	runtime := newStreamRuntime(ctrMgr)
	return NewServer(config, runtime)
}

type streamRuntime struct {
	containerMgr mgr.ContainerMgr
}

func newStreamRuntime(ctrMgr mgr.ContainerMgr) Runtime {
	return &streamRuntime{containerMgr: ctrMgr}
}

// Exec executes a command inside the container.
func (s *streamRuntime) Exec(containerID string, cmd []string, streamOpts *remotecommand.Options, streams *remotecommand.Streams) (uint32, error) {
	createConfig := &apitypes.ExecCreateConfig{
		Cmd:          cmd,
		AttachStdin:  streamOpts.Stdin,
		AttachStdout: streamOpts.Stdout,
		AttachStderr: streamOpts.Stderr,
		Tty:          streamOpts.TTY,
	}

	ctx := context.Background()

	execid, err := s.containerMgr.CreateExec(ctx, containerID, createConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to create exec for container %q: %v", containerID, err)
	}

	attachConfig := &mgr.AttachConfig{
		Streams:     streams,
		MuxDisabled: true,
	}

	err = s.containerMgr.StartExec(ctx, execid, attachConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to start exec for container %q: %v", containerID, err)
	}

	// TODO Find a better way instead of the dead loop
	var ei *apitypes.ContainerExecInspect
	for {
		ei, err = s.containerMgr.InspectExec(ctx, execid)
		if err != nil {
			return 0, fmt.Errorf("failed to inspect exec for container %q: %v", containerID, err)
		}
		// Loop until exec finished.
		if !ei.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return uint32(ei.ExitCode), nil
}

// Attach attaches to a running container.
func (s *streamRuntime) Attach(containerID string, streamOpts *remotecommand.Options, streams *remotecommand.Streams) error {
	attachConfig := &mgr.AttachConfig{
		Stdin:   streamOpts.Stdin,
		Stdout:  streamOpts.Stdout,
		Stderr:  streamOpts.Stderr,
		Streams: streams,
	}

	err := s.containerMgr.Attach(context.Background(), containerID, attachConfig)
	if err != nil {
		return fmt.Errorf("failed to attach to container %q: %v", containerID, err)
	}

	<-streams.StreamCh

	return nil
}

// PortForward forwards ports from a PodSandbox.
func (s *streamRuntime) PortForward(id string, port int32, stream io.ReadWriteCloser) error {
	sandbox, err := s.containerMgr.Get(context.Background(), id)
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
