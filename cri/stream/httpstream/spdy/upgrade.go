package spdy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/alibaba/pouch/cri/stream/httpstream"

	"github.com/sirupsen/logrus"
)

// HeaderSpdy31 is used to specify SPDY/3.1 as the stream protocol.
const HeaderSpdy31 = "SPDY/3.1"

// responseUpgrader knows how to upgrade HTTP responses. It
// implements the httpstream.ResponseUpgrader interface.
type responseUpgrader struct {
}

// connWrapper is used to wrap a hijacked connection and its bufio.Reader. All
// calls will be handled directly by the underlying net.Conn with the exception
// of Read and Close calls, which will consider data in the bufio.Reader. This
// ensures that data already inside the used bufio.Reader instance is also
// read.
type connWrapper struct {
	net.Conn
	closed    int32
	bufReader *bufio.Reader
}

func (w *connWrapper) Read(b []byte) (n int, err error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return 0, io.EOF
	}
	return w.bufReader.Read(b)
}

func (w *connWrapper) Close() error {
	err := w.Conn.Close()
	atomic.StoreInt32(&w.closed, 1)
	return err
}

// NewResponseUpgrader returns a new httpstream.ResponseUpgrader that is
// capable of upgrading HTTP responses using SPDY/3.1 via the
// spdystream package.
func NewResponseUpgrader() httpstream.ResponseUpgrader {
	return responseUpgrader{}
}

// UpgradeResponse upgrades an HTTP response to one that supports multiplexed
// streams. newStreamHandler will be called synchronously whenever the
// other end of the upgraded connection creates a new stream.
func (u responseUpgrader) UpgradeResponse(w http.ResponseWriter, req *http.Request, newStreamHandler httpstream.NewStreamHandler) httpstream.Connection {
	connectionHeader := strings.ToLower(req.Header.Get(httpstream.HeaderConnection))
	upgradeHeader := strings.ToLower(req.Header.Get(httpstream.HeaderUpgrade))
	if !strings.Contains(connectionHeader, strings.ToLower(httpstream.HeaderUpgrade)) || !strings.Contains(upgradeHeader, strings.ToLower(HeaderSpdy31)) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unable to upgrade: missing upgrade headers in request: %#v", req.Header)
		return nil
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "unable to upgrade: unable to hijack response")
		return nil
	}

	w.Header().Add(httpstream.HeaderConnection, httpstream.HeaderUpgrade)
	w.Header().Add(httpstream.HeaderUpgrade, HeaderSpdy31)
	w.WriteHeader(http.StatusSwitchingProtocols)

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		logrus.Errorf("unable to upgrade: error hijacking response: %v", err)
		return nil
	}

	connWithBuf := &connWrapper{Conn: conn, bufReader: bufrw.Reader}
	spdyConn, err := NewServerConnection(connWithBuf, newStreamHandler)
	if err != nil {
		logrus.Errorf("unable to upgrade: error creating SPDY server connection: %v", err)
		return nil
	}

	return spdyConn
}
