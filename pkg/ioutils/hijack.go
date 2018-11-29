package ioutils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/term"
	"github.com/sirupsen/logrus"
)

var defaultEscapeKeys = []byte{16, 17}

// HijackedIOStreamer represents io stream
type HijackedIOStreamer struct {
	Streams      Streams
	InputStream  io.ReadCloser
	OutputStream io.Writer

	Conn      net.Conn
	BufReader *bufio.Reader

	Tty        bool
	DetachKeys string
}

// Stream is used to streaming io
func (h *HijackedIOStreamer) Stream(ctx context.Context) error {
	restoreInput, err := h.setupInput()
	if err != nil {
		return fmt.Errorf("unable to setup input stream: %s", err)
	}

	defer restoreInput()

	outputDone := h.beginOutputStream(restoreInput)
	inputDone := h.beginInputStream(restoreInput)

	select {
	case err := <-outputDone:
		return err
	case err := <-inputDone:
		if err != nil { // if not detach keys, we should wait for outputDone
			if h.OutputStream != nil {
				select {
				case err := <-outputDone:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *HijackedIOStreamer) setupInput() (restore func(), err error) {
	if h.InputStream == nil || !h.Tty {
		return func() {}, nil
	}

	if err := setRawTerminal(h.Streams); err != nil {
		return nil, fmt.Errorf("unable to set IO streams as raw terminal: %s", err)
	}

	var restoreOnce sync.Once
	restore = func() {
		restoreOnce.Do(func() {
			restoreTerminal(h.Streams, h.InputStream)
		})
	}

	// Wrap the escape of detach keys
	escapeKeys := defaultEscapeKeys
	if h.DetachKeys != "" {
		customEscapeKeys, err := term.ToBytes(h.DetachKeys)
		if err != nil {
			logrus.Warnf("invalid detach escape keys, using default: %s", err)
		} else {
			escapeKeys = customEscapeKeys
		}
	}

	h.InputStream = ioutils.NewReadCloserWrapper(term.NewEscapeProxy(h.InputStream, escapeKeys), h.InputStream.Close)

	return restore, nil
}

func (h *HijackedIOStreamer) beginOutputStream(restoreInput func()) <-chan error {
	if h.OutputStream == nil {
		return nil
	}

	outputDone := make(chan error)
	go func() {
		var err error
		if h.OutputStream != nil {
			_, err = io.Copy(h.OutputStream, h.BufReader)
		}

		if h.Tty {
			restoreInput()
		}

		outputDone <- err
	}()

	return outputDone
}

func (h *HijackedIOStreamer) beginInputStream(restoreInput func()) <-chan error {
	inputDone := make(chan error)

	go func() {
		if h.InputStream != nil {
			_, err := io.Copy(h.Conn, h.InputStream)
			restoreInput()

			if _, ok := err.(term.EscapeError); ok {
				inputDone <- nil
				return
			}
		}

		if cw, ok := h.Conn.(CloseWriter); ok {
			cw.CloseWrite()
		}
	}()

	return inputDone
}

func setRawTerminal(streams Streams) error {
	if err := streams.In().SetRawTerminal(); err != nil {
		return err
	}
	return streams.Out().SetRawTerminal()
}

func restoreTerminal(streams Streams, in io.Closer) error {
	streams.In().RestoreTerminal()
	streams.Out().RestoreTerminal()
	return in.Close()
}
