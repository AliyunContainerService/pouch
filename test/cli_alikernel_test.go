package main

import (
	"runtime"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchAliKernelSuite is the test suite for AliOS specified features.
type PouchAliKernelSuite struct{}

func init() {
	check.Suite(&PouchAliKernelSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchAliKernelSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
	SkipIfFalse(c, environment.IsAliKernel)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchAliKernelSuite) TearDownTest(c *check.C) {
}

// TestAliKernelDiskQuotaWorks tests disk quota works on AliKernel.
func (suite *PouchAliKernelSuite) TestAliKernelDiskQuotaWorks(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	command.PouchRun("volume", "create", "--name", funcname, "-d", "local", "-o", "opt.size=1g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", funcname)

	command.PouchRun("run", "-d", "-v", funcname+":/mnt", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, funcname)

	expct := icmd.Expected{
		ExitCode: 0,
		Out:      "1.0G",
	}
	err := command.PouchRun("exec", funcname, "df", "-h").Compare(expct)
	c.Assert(err, check.IsNil)

	// generate a file larger than 1G should fail.
	expct = icmd.Expected{
		// TODO: Add exit code check when pouch exec return the exit code of process
		Error: "Disk quota exceeded",
	}
	cmd := "dd if=/dev/zero of=/mnt/test bs=1024k count=1500"
	err = command.PouchRun("exec", funcname, "sh", "-c", cmd).Compare(expct)
	c.Assert(err, check.IsNil)
}

// TestAliKernelDiskQuotaMultiWorks tests multi volume with different disk quota works on AliKernel.
func (suite *PouchAliKernelSuite) TestAliKernelDiskQuotaMultiWorks(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	name1 := funcname + "1"
	name2 := funcname + "2"

	command.PouchRun("volume", "create", "--name", name1, "-d", "local", "-o", "opt.size=2.2g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", name1)

	command.PouchRun("volume", "create", "--name", name2, "-d", "local", "-o", "opt.size=3.2g").Assert(c, icmd.Success)
	defer command.PouchRun("volume", "rm", name2)

	command.PouchRun("run", "-d", "-v", name1+":/mnt/test1", "-v", name2+":/mnt/test2", "--name", funcname, busyboxImage, "top").Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, funcname)

	{
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      "2.2G",
		}
		cmd := "df -h |grep test1"
		err := command.PouchRun("exec", funcname, "sh", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)
	}
	{
		expct := icmd.Expected{
			ExitCode: 0,
			Out:      "3.2G",
		}
		cmd := "df -h |grep test2"
		err := command.PouchRun("exec", funcname, "sh", "-c", cmd).Compare(expct)
		c.Assert(err, check.IsNil)
	}
}
