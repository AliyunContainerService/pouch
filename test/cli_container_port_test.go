package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchContainerPortSuite is the test suite for container port CLI.
type PouchContainerPortSuite struct{}

func init() {
	check.Suite(&PouchContainerPortSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchContainerPortSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	PullImage(c, busyboxImage)
}

// Test pouch port
func (suite *PouchContainerPortSuite) TestPouchPort(c *check.C) {
	testcase1 := "8000:8000"
	testcase2 := "10000:10000/udp"
	testcase3 := "127.0.0.1:5000:5000"

	name := "TestPouchPort"
	command.PouchRun("run",
		"--name", name, "-d",
		"-p", testcase1,
		"-p", testcase2,
		"-p", testcase3,
		busyboxImage,
		"sh", "-c", "sleep 10000").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	ret := command.PouchRun("port",
		name, "5000").Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "127.0.0.1:5000\n")

	ret = command.PouchRun("port",
		name, "8000").Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "0.0.0.0:8000\n")

	ret = command.PouchRun("port",
		name, "10000/udp").Assert(c, icmd.Success)
	c.Assert(ret.Stdout(), check.Equals, "0.0.0.0:10000\n")

	portBindingMap := map[string]string{
		"10000/udp": "0.0.0.0:10000",
		"5000/tcp":  "127.0.0.1:5000",
		"8000/tcp":  "0.0.0.0:8000",
	}
	// Test for only one arg.
	ret = command.PouchRun("port", name).Assert(c, icmd.Success)
	lines := strings.Split(ret.Stdout(), "\n")
	for _, l := range lines {
		outputs := strings.Split(l, "->")
		// filter out the last line
		if len(outputs) < 2 {
			continue
		}
		port := strings.TrimSpace(outputs[0])
		actualPortBinding := strings.TrimSpace(outputs[1])
		if expected, exist := portBindingMap[port]; exist {
			c.Assert(expected, check.Equals, actualPortBinding)
		} else {
			c.Errorf("Port %s not exists for container %s", port, name)
		}
	}
}
