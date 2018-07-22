package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	digest "github.com/opencontainers/go-digest"
)

// PouchCreateSuite is the test suite for create CLI.
type PouchCreateSuite struct{}

func init() {
	check.Suite(&PouchCreateSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCreateSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchCreateSuite) TearDownTest(c *check.C) {
}

// TestCreateName is to verify the correctness of creating container with specified name.
func (suite *PouchCreateSuite) TestCreateName(c *check.C) {
	name := "create-normal"
	res := command.PouchRun("create", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	// create command should add newline at the end of result
	digStr := strings.TrimSpace(res.Combined())
	c.Assert(res.Combined(), check.Equals, fmt.Sprintf("%s\n", digStr))

}

// TestCreateNameByImageID is to verify the correctness of creating contaier with specified name by image id.
func (suite *PouchCreateSuite) TestCreateNameByImageID(c *check.C) {
	name := "create-normal-by-image-id"

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)
	imageID := imagesListToKV(res.Combined())[busyboxImage][0]

	res = command.PouchRun("create", "--name", name, imageID)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	digHexStr := strings.TrimSpace(res.Combined())
	_, err := digest.Parse(fmt.Sprintf("%s:%s", digest.SHA256, digHexStr))
	c.Assert(err, check.IsNil)

}

// TestCreateDuplicateContainerName is to verify duplicate container names.
func (suite *PouchCreateSuite) TestCreateDuplicateContainerName(c *check.C) {
	name := "duplicate"

	res := command.PouchRun("create", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("create", "--name", name, busyboxImage)
	c.Assert(res.Stderr(), check.NotNil)

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
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// TestCreateWithTTY is to verify tty flag.
//
// TODO: pouch inspect should return tty info
func (suite *PouchCreateSuite) TestCreateWithTTY(c *check.C) {
	name := "TestCreateWithTTY"
	res := command.PouchRun("create", "-t", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
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
	defer DelContainerForceMultyTime(c, funcname)

	res.Assert(c, icmd.Success)
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
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}

// TestCreateWithLabels tries to test create a container with label.
func (suite *PouchCreateSuite) TestCreateWithLabels(c *check.C) {
	label := "abc=123"
	name := "create-label"

	res := command.PouchRun("create", "--name", name, "-l", label, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.Labels, check.NotNil)

	if result[0].Config.Labels["abc"] != "123" {
		c.Errorf("failed to set label: %s", label)
	}
}

// TestCreateWithSysctls tries to test create a container with sysctls.
func (suite *PouchCreateSuite) TestCreateWithSysctls(c *check.C) {
	sysctl := "net.ipv4.ip_forward=1"
	name := "create-sysctl"

	res := command.PouchRun("create", "--name", name, "--sysctl", sysctl, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.Sysctls, check.NotNil)

	if result[0].HostConfig.Sysctls["net.ipv4.ip_forward"] != "1" {
		c.Errorf("failed to set sysctl: %s", sysctl)
	}
}

// TestCreateWithAppArmor tries to test create a container with security option AppArmor.
func (suite *PouchCreateSuite) TestCreateWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "create-apparmor"

	res := command.PouchRun("create", "--name", name, "--security-opt", appArmor, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.SecurityOpt, check.NotNil)

	exist := false
	for _, opt := range result[0].HostConfig.SecurityOpt {
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
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.SecurityOpt, check.NotNil)

	exist := false
	for _, opt := range result[0].HostConfig.SecurityOpt {
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
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.CapAdd, check.NotNil)

	exist := false
	for _, cap := range result[0].HostConfig.CapAdd {
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
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.Privileged, check.Equals, true)
}

// TestCreateEnableLxcfs tries to test create a container with lxcfs.
func (suite *PouchCreateSuite) TestCreateEnableLxcfs(c *check.C) {
	name := "create-lxcfs"

	res := command.PouchRun("create", "--name", name, "--enableLxcfs=true", busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.EnableLxcfs, check.NotNil)

	if result[0].HostConfig.EnableLxcfs != true {
		c.Errorf("failed to set EnableLxcfs")
	}
}

// TestCreateWithEnv tests creating container with env
func (suite *PouchCreateSuite) TestCreateWithEnv(c *check.C) {
	name := "TestCreateWithEnv"

	res := command.PouchRun("create", "--name", name, "-e TEST=true", busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	ok := false
	for _, v := range result[0].Config.Env {
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
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(strings.TrimSpace(result[0].Config.WorkingDir), check.Equals, "/tmp/test")

	// TODO: check the work directory has been created.
}

// TestCreateWithUser tests creating container with user works.
func (suite *PouchCreateSuite) TestCreateWithUser(c *check.C) {
	name := "TestCreateWithUser"
	user := "1001"

	res := command.PouchRun("create", "--name", name, "--user", user, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].Config.User, check.Equals, user)
}

// TestCreateWithIntelRdt tests creating container with Intel Rdt.
func (suite *PouchCreateSuite) TestCreateWithIntelRdt(c *check.C) {
	name := "TestCreateWithIntelRdt"
	intelRdt := "L3:<cache_id0>=<cbm0>"

	res := command.PouchRun("create", "--name", name, "--intel-rdt-l3-cbm", intelRdt, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.IntelRdtL3Cbm, check.Equals, intelRdt)
}

// TestCreateWithAliOSMemoryOptions tests creating container with AliOS container isolation options.
func (suite *PouchCreateSuite) TestCreateWithAliOSMemoryOptions(c *check.C) {
	name := "TestCreateWithAliOSMemoryOptions"
	memoryWmarkRatio := "30"
	memoryExtra := "50"

	res := command.PouchRun("create", "--name", name, "--memory-wmark-ratio",
		memoryWmarkRatio, "--memory-extra", memoryExtra, "--memory-force-empty-ctl", "1",
		"--sche-lat-switch", "1", busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(*result[0].HostConfig.MemoryWmarkRatio, check.Equals, int64(30))
	c.Assert(*result[0].HostConfig.MemoryExtra, check.Equals, int64(50))
	c.Assert(result[0].HostConfig.MemoryForceEmptyCtl, check.Equals, int64(1))
	c.Assert(result[0].HostConfig.ScheLatSwitch, check.Equals, int64(1))
}

// TestCreateWithOOMOption tests creating container with oom options.
func (suite *PouchCreateSuite) TestCreateWithOOMOption(c *check.C) {
	name := "TestCreateWithOOMOption"
	oomScore := "100"

	res := command.PouchRun("create", "--name", name, "--oom-score-adj", oomScore,
		"--oom-kill-disable", busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.OomScoreAdj, check.Equals, int64(100))
	c.Assert(*result[0].HostConfig.OomKillDisable, check.Equals, true)
}

// TestCreateWithAnnotation tests creating container with annotation.
func (suite *PouchCreateSuite) TestCreateWithAnnotation(c *check.C) {
	cname := "TestCreateWithAnnotation"
	res := command.PouchRun("create", "--annotation", "a=b", "--annotation", "foo=bar",
		"--name", cname, busyboxImage)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// kv in map not in order.
	var annotationSlice []string
	for k, v := range result[0].Config.SpecAnnotation {
		annotationSlice = append(annotationSlice, fmt.Sprintf("%s=%s", k, v))
	}
	annotationStr := strings.Join(annotationSlice, " ")

	c.Assert(util.PartialEqual(annotationStr, "a=b"), check.IsNil)
	c.Assert(util.PartialEqual(annotationStr, "foo=bar"), check.IsNil)
}

// TestCreateWithUlimit tests creating container with annotation.
func (suite *PouchCreateSuite) TestCreateWithUlimit(c *check.C) {
	cname := "TestCreateWithUlimit"
	res := command.PouchRun("create", "--ulimit", "nproc=21", "--name", cname, busyboxImage)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	ul := result[0].HostConfig.Ulimits[0]
	c.Assert(ul.Name, check.Equals, "nproc")
	c.Assert(int(ul.Hard), check.Equals, 21)
	c.Assert(int(ul.Soft), check.Equals, 21)
}

// TestCreateWithPidsLimit tests running container with --pids-limit flag.
func (suite *PouchRunSuite) TestCreateWithPidsLimit(c *check.C) {
	cname := "TestCreateWithPidsLimit"
	res := command.PouchRun("create", "--pids-limit", "10", "--name", cname, busyboxImage)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	pl := result[0].HostConfig.PidsLimit
	c.Assert(int(pl), check.Equals, 10)
}

// TestCreateWithNonExistImage tests running container with image not exist.
func (suite *PouchRunSuite) TestCreateWithNonExistImage(c *check.C) {
	cname := "TestCreateWithNonExistImage"
	// we should use a non-used image, since containerd not remove image immediately.
	image := "docker.io/library/alpine"
	res := command.PouchRun("create", cname, image)
	res.Assert(c, icmd.Success)
}
