package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/gotestyourself/gotestyourself/icmd"

	"github.com/go-check/check"
)

// PouchRunSyslogSuite is the test suite for run CLI.
type PouchRunSyslogSuite struct{}

func init() {
	check.Suite(&PouchRunSyslogSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunSyslogSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

func (suite *PouchRunSyslogSuite) TestRunRFC5424MicroSeq(c *check.C) {
	msgCh := make(chan string)
	addr, conn := suite.startTCPServer(c, msgCh)
	defer conn.Close()

	type tCase struct {
		env         string // for container
		optTag      string
		optEnv      string
		expectedTag string
	}

	cname := "test-syslog-Basic"
	for i, tc := range []tCase{
		{
			env:         "POUCH_VERSION=ga",
			optTag:      "{{.POUCH_VERSION}}",
			optEnv:      "POUCH_VERSION",
			expectedTag: "ga",
		},
		{
			env:         "POUCH_BUILD=unknow",
			optTag:      "{{.POUCH_VERSION}}",
			optEnv:      "POUCH_BUILD",
			expectedTag: "<no value>",
		},
	} {
		name := fmt.Sprintf("%s-%d", cname, i+1)

		command.PouchRun("run", "-d",
			"--name", name,
			"--log-driver", "syslog",
			"--log-opt", "syslog-address=tcp://"+addr,
			"--log-opt", fmt.Sprintf("tag={{with .ExtraAttributes nil}}%s{{end}}", tc.optTag),
			"--log-opt", fmt.Sprintf("env=%s", tc.optEnv),
			"--log-opt", "syslog-format=rfc5424micro-seq",
			"--env", tc.env,
			busyboxImage, "echo", name,
		).Assert(c, icmd.Success)
		defer DelContainerForceMultyTime(c, name)

		// rfc5424micro-seq will has the suffix template like "{{tag}} - {{content}}
		c.Assert(suite.checkMessage(fmt.Sprintf("%s - %s\n", tc.expectedTag, name), msgCh), check.IsNil)
	}
}

func (suite *PouchRunSyslogSuite) checkMessage(expected string, msgCh <-chan string) error {
	var (
		msg string
		ok  bool
	)

	tc := time.NewTimer(1000 * time.Millisecond)
	defer tc.Stop()

	select {
	case msg, ok = <-msgCh:
		if !ok {
			return fmt.Errorf("failed to get message from msgCh")
		}
	case <-tc.C:
		return fmt.Errorf("failed to get message by timeout")
	}

	if !strings.HasSuffix(msg, expected) {
		return fmt.Errorf("expected has suffix %s, but got %s", expected, msg)
	}
	return nil
}

func (suite *PouchRunSyslogSuite) startTCPServer(t testingTB, msgCh chan<- string) (addr string, conn io.Closer) {
	var (
		li  net.Listener
		err error
	)

	// 127.0.0.1:0 will use random available port
	addr = "127.0.0.1:0"
	li, err = net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen on %s: %v", addr, err)
	}

	addr = li.Addr().String()
	conn = li

	go func() {
		for {
			var c net.Conn
			var err error

			if c, err = li.Accept(); err != nil {
				return
			}

			go func(c net.Conn) {
				c.SetReadDeadline(time.Now().Add(5 * time.Second))
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
