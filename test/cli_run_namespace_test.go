package main

import (
	"os"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// TestRunWithHostPidNS tests running container in host pid namespace
func (suite *PouchRunSuite) TestRunWithHostPidNS(c *check.C) {
	name := "TestRunWithHostPidNS"
	command.PouchRun("run", "-d", "--name", name, "--pid", "host", busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	ret := command.PouchRun("exec", name, "pidof", "top")
	ret.Assert(c, icmd.Success)

	if strings.TrimSpace(ret.Stdout()) == "1" {
		c.Fatal("container in host pid namespace can not get init pid as pid 1")
	}
}

// TestRunWithHostUTSNS tests running container in host uts namespace
func (suite *PouchRunSuite) TestRunWithHostUTSNS(c *check.C) {
	name := "TestRunWithHostUTSNS"
	command.PouchRun("run", "-d", "--name", name, "--uts", "host", busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)
	ret := command.PouchRun("exec", name, "hostname")
	ret.Assert(c, icmd.Success)

	hostname, err := os.Hostname()
	c.Assert(err, check.IsNil)

	if strings.TrimSpace(ret.Stdout()) != hostname {
		c.Fatalf("hostname %s should the same with host %s", ret.Stdout(), hostname)
	}
}

// TestRunWithHostIPCNS tests running container in host uts namespace can successfully run.
func (suite *PouchRunSuite) TestRunWithHostIPCNS(c *check.C) {
	name := "TestRunWithHostIPCNS"
	command.PouchRun("run", "-d", "--name", name, "--ipc", "host", busyboxImage, "top").Assert(c, icmd.Success)
	DelContainerForceMultyTime(c, name)
}

// TestRunWithInvalidNS tests running container with invalid namespace value should fail.
func (suite *PouchRunSuite) TestRunWithInvalidNS(c *check.C) {
	name := "TestRunWithInvalidPidNS"
	ret := command.PouchRun("run", "-d", "--name", name, "--pid", "invalid", busyboxImage, "top")
	c.Assert(util.PartialEqual(ret.Stderr(), "invalid pid namespace mode"), check.IsNil)
	DelContainerForceMultyTime(c, name)

	name = "TestRunWithInvalidUTSNS"
	ret = command.PouchRun("run", "-d", "--name", name, "--uts", "invalid", busyboxImage, "top")
	c.Assert(util.PartialEqual(ret.Stderr(), "invalid uts namespace mode"), check.IsNil)
	DelContainerForceMultyTime(c, name)

	name = "TestRunWithInvalidIPCNS"
	ret = command.PouchRun("run", "-d", "--name", name, "--ipc", "invalid", busyboxImage, "top")
	c.Assert(util.PartialEqual(ret.Stderr(), "invalid ipc namespace mode"), check.IsNil)
	DelContainerForceMultyTime(c, name)
}
