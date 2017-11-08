package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// StartCommand use to implement 'start' command, it start a container.
type StartCommand struct {
	baseCommand
	attach bool
	stdin  bool
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "start [container]",
		Short: "Start a created container",
		Args:  cobra.MinimumNArgs(1),
	}

	s.cmd.Flags().BoolVarP(&s.attach, "attach", "a", false, "attach the container's io or not")
	s.cmd.Flags().BoolVarP(&s.stdin, "interactive", "i", false, "attach container's stdin")
}

// Run is the entry of start command.
func (s *StartCommand) Run(args []string) {
	container := args[0]

	// attach to io.
	var wait chan struct{}
	if s.attach || s.stdin {
		in, out, err := setRawMode(s.stdin, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set raw mode")
			return
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		conn, br, err := attachContainer(s.cli, container, s.stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to attach container: %v \n", err)
			return
		}

		wait = make(chan struct{})
		go func() {
			io.Copy(os.Stdout, br)
			close(wait)
		}()
		go func() {
			io.Copy(conn, os.Stdin)
			close(wait)
		}()
	}

	// start container.
	path := fmt.Sprintf("/containers/%s/start", container)

	req, err := s.cli.NewPostRequest(path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new client: %v\n", err)
		return
	}

	response := req.Send()
	if err := response.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to request: %v \n", err)
		return
	}
	defer response.Close()

	// wait the io to finish.
	if s.attach || s.stdin {
		<-wait
	}
}

func attachContainer(c *Cli, name string, stdin bool) (net.Conn, *bufio.Reader, error) {
	var path string
	if stdin {
		path = fmt.Sprintf("/containers/%s/attach?stdin=1", name)
	} else {
		path = fmt.Sprintf("/containers/%s/attach?stdin=0", name)
	}
	req, err := c.NewPostRequest(path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.req.Header["Content-Type"] = []string{"text/plain"}
	req.req.Header.Set("Connection", "Upgrade")
	req.req.Header.Set("Upgrade", "tcp")

	var conn net.Conn

	if strings.HasPrefix(c.host, "unix://") {
		req.req.Host = strings.TrimPrefix(c.host, "unix://")
		if conn, err = net.DialTimeout("unix", req.req.Host, time.Second*10); err != nil {
			return nil, nil, err
		}
	} else {
		req.req.Host = strings.TrimPrefix(c.host, "tcp://")
		if conn, err = net.DialTimeout("tcp", req.req.Host, time.Second*10); err != nil {
			return nil, nil, err
		}
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	clientconn := httputil.NewClientConn(conn, nil)
	defer clientconn.Close()

	if _, err := clientconn.Do(req.req); err != nil {
		return nil, nil, err
	}

	rwc, br := clientconn.Hijack()
	return rwc, br, nil
}

func setRawMode(stdin, stdout bool) (*terminal.State, *terminal.State, error) {
	var (
		in  *terminal.State
		out *terminal.State
		err error
	)

	if stdin {
		if in, err = terminal.MakeRaw(0); err != nil {
			return nil, nil, err
		}
	}
	if stdout {
		if out, err = terminal.MakeRaw(1); err != nil {
			return nil, nil, err
		}
	}

	return in, out, nil
}

func restoreMode(in, out *terminal.State) error {
	if in != nil {
		if err := terminal.Restore(0, in); err != nil {
			return err
		}
	}
	if out != nil {
		if err := terminal.Restore(1, out); err != nil {
			return err
		}
	}
	return nil
}
