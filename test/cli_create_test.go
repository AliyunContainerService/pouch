package main

import (
	"encoding/json"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchCreateSuite is the test suite fo help CLI.
type PouchCreateSuite struct{}

func init() {
	check.Suite(&PouchCreateSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCreateSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	command.PouchRun("pull", busyboxImage).Assert(c, icmd.Success)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchCreateSuite) TearDownTest(c *check.C) {
}

// TestCreateName is to verify the correctness of creating contaier with specified name.
func (suite *PouchCreateSuite) TestCreateName(c *check.C) {
	name := "create-normal"
	res := command.PouchRun("create", "--name", name, busyboxImage)

	res.Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}

	defer DelContainerForceMultyTime(c, name)
}

// TestCreateNameByImageID is to verify the correctness of creating contaier with specified name by image id.
func (suite *PouchCreateSuite) TestCreateNameByImageID(c *check.C) {
	name := "create-normal-by-image-id"

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[busyboxImage][0]

	res = command.PouchRun("create", "--name", name, imageID)

	res.Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s expected %s\n", out, name)
	}

	DelContainerForceMultyTime(c, name)
}

// TestCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestCreateDuplicateContainerName(c *check.C) {
	name := "duplicate"

	res := command.PouchRun("create", "--name", name, busyboxImage)
	res.Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, name)

	res = command.PouchRun("create", "--name", name, busyboxImage)
	c.Assert(res.Error, check.NotNil)

	if out := res.Combined(); !strings.Contains(out, "already exist") {
		c.Fatalf("unexpected output %s expected already exist\n", out)
	}
}

// TestCreateWithArgs is to verify args.
//
// TODO: pouch inspect should return args info
func (suite *PouchCreateSuite) TestCreateWithArgs(c *check.C) {
	name := "TestCreateWithArgs"
	res := command.PouchRun("create", "--name", name, busyboxImage, "/bin/ls")
	res.Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, name)
}

// TestCreateWithTTY is to verify tty flag.
//
// TODO: pouch inspect should return tty info
func (suite *PouchCreateSuite) TestCreateWithTTY(c *check.C) {
	name := "TestCreateWithTTY"
	res := command.PouchRun("create", "-t", "--name", name, busyboxImage)
	res.Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, name)
}

// TestPouchCreateVolume is to verify volume flag.
//
// TODO: pouch inspect should return volume info to check
func (suite *PouchCreateSuite) TestPouchCreateVolume(c *check.C) {
	pc, _, _, _ := runtime.Caller(0)
	tmpname := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	var funcname string
	for i := range tmpname {
		funcname = tmpname[i]
	}

	res := command.PouchRun("create", "-v /tmp:/tmp", "--name", funcname, busyboxImage)
	res.Assert(c, icmd.Success)

	defer DelContainerForceMultyTime(c, funcname)
}

// TestCreateInWrongWay tries to run create in wrong way.
func (suite *PouchCreateSuite) TestCreateInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("create", tc.args)
		c.Assert(res.Error, check.NotNil, check.Commentf(tc.name))
	}
}

// TestCreateWithLabels tries to test create a container with label.
func (suite *PouchCreateSuite) TestCreateWithLabels(c *check.C) {
	label := "abc=123"
	name := "create-label"

	res := command.PouchRun("create", "--name", name, "-l", label, busyboxImage)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.Labels, check.NotNil)

	if result.Config.Labels["abc"] != "123" {
		c.Errorf("failed to set label: %s", label)
	}
}

// TestCreateWithSysctls tries to test create a container with sysctls.
func (suite *PouchCreateSuite) TestCreateWithSysctls(c *check.C) {
	sysctl := "net.ipv4.ip_forward=1"
	name := "create-sysctl"

	res := command.PouchRun("create", "--name", name, "--sysctl", sysctl, busyboxImage)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.Sysctls, check.NotNil)

	if result.HostConfig.Sysctls["net.ipv4.ip_forward"] != "1" {
		c.Errorf("failed to set sysctl: %s", sysctl)
	}
}

// TestCreateWithAppArmor tries to test create a container with security option AppArmor.
func (suite *PouchCreateSuite) TestCreateWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "create-apparmor"

	res := command.PouchRun("create", "--name", name, "--security-opt", appArmor, busyboxImage)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.SecurityOpt, check.NotNil)

	exist := false
	for _, opt := range result.HostConfig.SecurityOpt {
		if opt == appArmor {
			exist = true
		}
	}
	if !exist {
		c.Errorf("failed to set AppArmor in security-opt")
	}
}

// TestCreateWithSeccomp tries to test create a container with security option seccomp.
func (suite *PouchCreateSuite) TestCreateWithSeccomp(c *check.C) {
	seccomp := "seccomp=unconfined"
	name := "create-seccomp"

	res := command.PouchRun("create", "--name", name, "--security-opt", seccomp, busyboxImage)
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.SecurityOpt, check.NotNil)

	exist := false
	for _, opt := range result.HostConfig.SecurityOpt {
		if opt == seccomp {
			exist = true
		}
	}
	if !exist {
		c.Errorf("failed to set seccomp in security-opt")
	}
}

// TestCreateWithCapability tries to test create a container with capability.
func (suite *PouchCreateSuite) TestCreateWithCapability(c *check.C) {
	capability := "NET_ADMIN"
	name := "create-capability"

	res := command.PouchRun("create", "--name", name, "--cap-add", capability, busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.CapAdd, check.NotNil)

	exist := false
	for _, cap := range result.HostConfig.CapAdd {
		if cap == capability {
			exist = true
		}
	}
	if !exist {
		c.Errorf("failed to set capability")
	}
}

// TestCreateWithPrivilege tries to test create a container with privilege.
func (suite *PouchCreateSuite) TestCreateWithPrivilege(c *check.C) {
	name := "create-privilege"

	res := command.PouchRun("create", "--name", name, "--privileged", busyboxImage, "brctl", "addbr", "foobar")
	res.Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.Privileged, check.Equals, true)
}

// TestCreateEnableLxcfs tries to test create a container with lxcfs.
func (suite *PouchCreateSuite) TestCreateEnableLxcfs(c *check.C) {
	name := "create-lxcfs"

	res := command.PouchRun("create", "--name", name, "--enableLxcfs=true", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.EnableLxcfs, check.NotNil)

	if result.HostConfig.EnableLxcfs != true {
		c.Errorf("failed to set EnableLxcfs")
	}
}

// TestCreateWithEnv tests creating container with env
func (suite *PouchCreateSuite) TestCreateWithEnv(c *check.C) {
	name := "TestCreateWithEnv"

	res := command.PouchRun("create", "--name", name, "-e TEST=true", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	ok := false
	for _, v := range result.Config.Env {
		if strings.Contains(v, "TEST=true") {
			ok = true
		}
	}
	c.Assert(ok, check.Equals, true)
}

// TestCreateWithWorkDir tests creating container with a workdir works.
func (suite *PouchCreateSuite) TestCreateWithWorkDir(c *check.C) {
	name := "TestCreateWithWorkDir"

	res := command.PouchRun("create", "--name", name, "-w /tmp/test", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(strings.TrimSpace(result.Config.WorkingDir), check.Equals, "/tmp/test")

	// TODO: check the work directory has been created.
}

// TestCreateWithUser tests creating container with user works.
func (suite *PouchCreateSuite) TestCreateWithUser(c *check.C) {
	name := "TestCreateWithUser"
	user := "1001"

	res := command.PouchRun("create", "--name", name, "--user", user, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.Config.User, check.Equals, user)
}

// TestCreateWithIntelRdt tests creating container with Intel Rdt.
func (suite *PouchCreateSuite) TestCreateWithIntelRdt(c *check.C) {
	name := "TestCreateWithIntelRdt"
	intelRdt := "L3:<cache_id0>=<cbm0>"

	res := command.PouchRun("create", "--name", name, "--intel-rdt-l3-cbm", intelRdt, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result.HostConfig.IntelRdtL3Cbm, check.Equals, intelRdt)
}

// TestCreateWithAliOSMemoryOptions tests creating container with AliOS container isolation options.
func (suite *PouchCreateSuite) TestCreateWithAliOSMemoryOptions(c *check.C) {
	name := "TestCreateWithAliOSMemoryOptions"
	memoryWmarkRatio := "30"
	memoryExtra := "50"

	res := command.PouchRun("create", "--name", name, "--memory-wmark-ratio", memoryWmarkRatio, "--memory-extra", memoryExtra, "--memory-force-empty-ctl", "1", "--sche-lat-switch", "1", busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := &types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(*result.HostConfig.MemoryWmarkRatio, check.Equals, int64(30))
	c.Assert(*result.HostConfig.MemoryExtra, check.Equals, int64(50))
	c.Assert(result.HostConfig.MemoryForceEmptyCtl, check.Equals, int64(1))
	c.Assert(result.HostConfig.ScheLatSwitch, check.Equals, int64(1))
}
