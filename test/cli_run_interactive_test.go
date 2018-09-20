package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunInteractiveSuite is the test suite for run CLI.
type PouchRunInteractiveSuite struct{}

func init() {
	check.Suite(&PouchRunInteractiveSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunInteractiveSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestRunInteractive test "-i" option works.
//
// FIXME(fuweid): when we start container, we need to save the pack into cache
// before the start task. otherwise, the c.watch.get(containerID) will return
// 404 and fail to stop container.
// before we fix this issue, we set it flaky case.
func (suite *PouchRunInteractiveSuite) TestRunInteractive(c *check.C) {
	c.Skip("skip flaky test")

	name := "TestRunInteractiveContainer"
	DelContainerForceMultyTime(c, name)

	// use pipe to act interactive
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	defer stdoutW.Close()

	// need to start command before write
	cmd := command.PouchCmd("run",
		"-i", "--net", "none",
		"--name", name, busyboxImage, "cat")
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	res := icmd.StartCmd(cmd)

	// send bye to cat
	//
	// FIXME(fuweid): when we remove ringbuffer, we can remove the
	// stdoutW, stdoutR pipe.
	fmt.Fprintf(stdinW, "hi\n")
	got, _, err := bufio.NewReader(stdoutR).ReadLine()
	c.Assert(err, check.IsNil)
	c.Assert(string(got), check.Equals, "hi")

	// send the EOF to stdin
	//
	// FIXME(fuweid): remove the timeout when we remove ringbuffer
	stdinW.Close()
	res = icmd.WaitOnCmd(1*time.Second, res)
	res.Assert(c, icmd.Success)

	// NOTE: container must be stopped.
	{
		output := command.PouchRun("inspect", name).Stdout()
		result := []types.ContainerJSON{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			c.Errorf("failed to decode inspect output: %v", err)
		}
		c.Assert(string(result[0].State.Status), check.Equals, "stopped")
	}
}
