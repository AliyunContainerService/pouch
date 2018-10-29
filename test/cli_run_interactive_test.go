package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
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
func (suite *PouchRunInteractiveSuite) TestRunInteractive(c *check.C) {
	name := "TestRunInteractiveContainer"
	defer DelContainerForceMultyTime(c, name)

	// use pipe to act interactive
	stdinR, stdinW := io.Pipe()
	stdoutBuf := bytes.NewBuffer(nil)
	defer stdinR.Close()

	// need to start command before write
	cmd := command.PouchCmd("run",
		"-i", "--net", "none",
		"--name", name, busyboxImage, "cat")
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutBuf
	res := icmd.StartCmd(cmd)

	// send bye to cat
	content := "byte\n"
	fmt.Fprintf(stdinW, content)
	stdinW.Close()

	res = icmd.WaitOnCmd(100*time.Second, res)
	res.Assert(c, icmd.Success)

	// NOTE: container must be exited.
	{
		output := command.PouchRun("inspect", name).Stdout()
		result := []types.ContainerJSON{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			c.Errorf("failed to decode inspect output: %v", err)
		}
		c.Assert(string(result[0].State.Status), check.Equals, "exited")
	}

	c.Assert(stdoutBuf.String(), check.Equals, content)
}

// TestRunForCloseIO test CloseIO works.
func (suite *PouchRunInteractiveSuite) TestRunForCloseIO(c *check.C) {
	name := "TestRunForCloseIO"
	defer DelContainerForceMultyTime(c, name)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cmdLine := fmt.Sprintf("echo 1 | %s run -i %s sh -c 'cat && echo hello'", environment.PouchBinary, busyboxImage)
	out, err := exec.CommandContext(ctx, "bash", "-c", cmdLine).Output()
	c.Assert(err, check.IsNil)
	c.Assert(string(out), check.Equals, "1\nhello\n")
}
