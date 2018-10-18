package containerio

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
)

func init() {
	Register(func() Backend {
		return &hijackConn{}
	})
}

type hijackConn struct {
	hijack      http.Hijacker
	conn        net.Conn
	closed      bool
	muxDisabled bool
}

func createHijackConn() Backend {
	return &hijackConn{}
}

func (h *hijackConn) Name() string {
	return "hijack"
}

func (h *hijackConn) Init(opt *Option) error {
	conn, _, err := opt.hijack.Hijack()
	if err != nil {
		return err
	}

	// set raw mode
	conn.Write([]byte{})

	if opt.hijackUpgrade {
		fmt.Fprintf(conn, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	} else {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
	}

	h.hijack = opt.hijack
	h.conn = conn
	h.muxDisabled = opt.muxDisabled
	return nil
}

func (h *hijackConn) Out() io.Writer {
	if !h.muxDisabled {
		return stdcopy.NewStdWriter(h, stdcopy.Stdout)
	}
	return h
}

func (h *hijackConn) In() io.Reader {
	return h
}

func (h *hijackConn) Err() io.Writer {
	if !h.muxDisabled {
		return stdcopy.NewStdWriter(h, stdcopy.Stderr)
	}
	return h
}

func (h *hijackConn) Close() error {
	if h.closed {
		return nil
	}
	h.closed = true
	return h.conn.Close()
}

func (h *hijackConn) Write(data []byte) (int, error) {
	return h.conn.Write(data)
}

func (h *hijackConn) Read(p []byte) (int, error) {
	return h.conn.Read(p)
}
