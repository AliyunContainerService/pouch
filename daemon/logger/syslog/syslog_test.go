package syslog

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"

	"github.com/RackSec/srslog"
)

type testingTB interface {
	Fatalf(format string, args ...interface{})
}

func TestParseOptions(t *testing.T) {
	info := logger.Info{
		LogConfig: map[string]string{
			"syslog-address":  "tcp://localhost:8080",
			"syslog-facility": "daemon",
			"syslog-format":   "rfc3164",
			"tag":             "{{.FullID}}",
		},
		ContainerID: "container-20180610",
	}

	opts, err := parseOptions(info)
	if err != nil {
		t.Fatalf("failed to parse options: %v", err)
	}

	// check formatter and framer
	if !isSameFunc(opts.formatter, srslog.RFC3164Formatter) || !isSameFunc(opts.framer, srslog.DefaultFramer) {
		t.Fatalf("expect formatter(%v) & framer(%v), but got formatter(%v) & framer(%v)",
			srslog.RFC3164Formatter, srslog.DefaultFramer, opts.formatter, opts.framer,
		)
	}

	// check tag
	if expected := info.ContainerID; opts.tag != expected {
		t.Fatalf("expect tag(%v), but got tag(%v)", expected, opts.tag)
	}

	// check priority
	if expected := srslog.LOG_DAEMON; opts.priority != expected {
		t.Fatalf("expect priority(%v), but got priority(%v)", expected, opts.priority)
	}

	// check proto and address
	if proto, addr := "tcp", "localhost:8080"; opts.proto != proto || opts.address != addr {
		t.Fatalf("expect proto(%v) & address(%v), but got proto(%v) & address(%v)",
			proto, addr, opts.proto, opts.address,
		)
	}
}

func TestConnectUnixSocket(t *testing.T) {
	msgCh := make(chan string)

	addr, conn, wg := startStreamServer("unix", 2, msgCh)
	defer func() {
		conn.Close()
		wg.Wait()
	}()

	info := logger.Info{
		LogConfig: map[string]string{
			"syslog-address": "unix://" + addr,
		},
	}

	slog, err := NewSyslog(info)
	if err != nil {
		t.Fatalf("failed to create Syslog: %v", err)
	}

	msg := "hi"
	if err := slog.logInfo(msg); err != nil {
		t.Fatalf("failed to logInfo: %v", err)
	}
	checkUnixFormatterMessage(t, srslog.LOG_INFO|srslog.LOG_DAEMON, slog.opt.tag, msg, msgCh)
}

func TestLazyAndRetryConnect(t *testing.T) {
	msgCh := make(chan string)

	addr, conn, wg := startStreamServer("tcp", 3, msgCh)
	defer func() {
		conn.Close()
		wg.Wait()
	}()

	info := logger.Info{
		LogConfig: map[string]string{
			"syslog-address": "tcp://" + addr,
		},
	}
	slog, err := NewSyslog(info)
	if err != nil {
		t.Fatalf("failed to create Syslog: %v", err)
	}

	// try to connect to the stream server
	{
		msg := "hi"
		if err := slog.logInfo(msg); err != nil {
			t.Fatalf("failed to logInfo: %v", err)
		}
		checkUnixFormatterMessage(t, srslog.LOG_INFO|srslog.LOG_DAEMON, slog.opt.tag, msg, msgCh)

		msg = "oops"
		if err := slog.logError(msg); err != nil {
			t.Fatalf("failed to logError: %v", err)
		}
		checkUnixFormatterMessage(t, srslog.LOG_ERR|srslog.LOG_DAEMON, slog.opt.tag, msg, msgCh)
	}

	// stop the connection and retry to connect to the stream server
	slog.getConn().close()
	{
		msg := "again+log-alert"
		if _, err := slog.writeAndRetry(srslog.LOG_ALERT, msg); err != nil {
			t.Fatalf("should reconnect, but got unexpected error here: %v", err)
		}
		checkUnixFormatterMessage(t, srslog.LOG_DAEMON|srslog.LOG_ALERT, slog.opt.tag, msg, msgCh)
	}
}

func checkUnixFormatterMessage(t testingTB, p Priority, tag, content string, msgCh <-chan string) {
	var (
		msg string
		ok  bool
	)

	tc := time.NewTimer(1000 * time.Millisecond)
	defer tc.Stop()

	select {
	case msg, ok = <-msgCh:
		if !ok {
			t.Fatalf("failed to get message from msgCh``")
		}
	case <-tc.C:
		t.Fatalf("failed to get message by timeout")
	}

	var (
		prefixTmpl = fmt.Sprintf("<%d>", p)
		suffixTmpl = fmt.Sprintf("%s[%d]: %s\n", tag, os.Getpid(), content)
	)

	if !strings.HasPrefix(msg, prefixTmpl) {
		t.Fatalf("should contains prefix %s, but got %v", prefixTmpl, msg)
	}

	if !strings.HasSuffix(msg, suffixTmpl) {
		t.Fatalf("should contains suffix %s, but got %v", suffixTmpl, msg)
	}
}

func TestTLSDialer(t *testing.T) {
	msgCh := make(chan string)

	addr, conn, _ := startStreamServer("tcp+tls", 3, msgCh)
	defer conn.Close()

	pool := x509.NewCertPool()
	cert, err := ioutil.ReadFile("test/ca.pem")
	if err != nil {
		t.Errorf("failed to read cert file: %v", err)
	}

	pool.AppendCertsFromPEM(cert)
	config := tls.Config{
		RootCAs: pool,
	}

	_, _, err = tlsDialer(addr, &config)
	if err != nil {
		t.Errorf("failed to dial: %v", err)
	}
}

// startStreamServer starts stream server which holds the connection after timeout.
func startStreamServer(proto string, readTimeout int, msgCh chan<- string) (addr string, conn io.Closer, drainWg *sync.WaitGroup) {
	if proto != "tcp" && proto != "tcp+tls" && proto != "unix" {
		log.Fatalf("not support %s", proto)
	}

	var (
		li   net.Listener
		err  error
		cert tls.Certificate
	)

	// 127.0.0.1:0 will use random available port
	addr = "127.0.0.1:0"
	if proto == "unix" {
		addr = randomUnixSocketName()
	}

	if proto == "tcp+tls" {
		cert, err = tls.LoadX509KeyPair("test/ca.pem", "test/ca-key.pem")
		if err != nil {
			log.Fatalf("failed to load TLS keypair: %v", err)
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}
		li, err = tls.Listen("tcp", addr, &config)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", addr, err)
		}
	} else {
		li, err = net.Listen(proto, addr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", addr, err)
		}
	}

	addr = li.Addr().String()
	conn = li
	drainWg = new(sync.WaitGroup)

	go func() {
		for {
			var c net.Conn
			var err error

			if c, err = li.Accept(); err != nil {
				return
			}

			drainWg.Add(1)
			go func(c net.Conn) {
				defer drainWg.Done()

				c.SetReadDeadline(time.Now().Add((time.Duration(readTimeout) * time.Second)))
				b := bufio.NewReader(c)

				for {
					s, err := b.ReadString('\n')
					if err != nil {
						break
					}
					msgCh <- s
				}
				c.Close()
			}(c)
		}
	}()
	return
}

// randomUnixSocketName uses TempFile to create random file name.
func randomUnixSocketName() (name string) {
	f, err := ioutil.TempFile("", "syslog-test-")
	if err != nil {
		log.Fatal("TempFile: ", err)
	}

	name = f.Name()
	f.Close()
	os.Remove(name)
	return
}
