package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/opencontainers/go-digest"
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

// TestCreateNameByImageID is to verify the correctness of creating container with specified name by image id.
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
	funcname := "TestPouchCreateVolume"

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
	name := "test-create-with-labels"

	expectedLabels := map[string]string{
		"abc": "123",
		"ABC": "123,456,789",
	}

	res := command.PouchRun("create",
		"--name", name,
		"-l", "abc=123",
		"-l", "ABC=123,456,789",
		busyboxImage,
	)

	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", name).Stdout()

	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	c.Assert(result[0].Config.Labels, check.NotNil)
	if !reflect.DeepEqual(result[0].Config.Labels, expectedLabels) {
		c.Fatalf("expected %v, but got %v", expectedLabels, result[0].Config.Labels)
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

	privileged, err := inspectFilter(name, ".HostConfig.Privileged")
	c.Assert(err, check.IsNil)
	c.Assert(privileged, check.Equals, "true")
}

// TestCreateEnableLxcfs tries to test create a container with lxcfs.
func (suite *PouchCreateSuite) TestCreateEnableLxcfs(c *check.C) {
	name := "create-lxcfs"

	res := command.PouchRun("create", "--name", name, "--enableLxcfs=true", busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	enableLxcfs, err := inspectFilter(name, ".HostConfig.EnableLxcfs")
	c.Assert(err, check.IsNil)
	c.Assert(enableLxcfs, check.Equals, "true")
}

// TestCreateWithEnv tests creating container with env
func (suite *PouchCreateSuite) TestCreateWithEnv(c *check.C) {
	name := "TestCreateWithEnv"

	env1 := "TEST1=true"
	env2 := "TEST2="    // should be in inspect result as TEST2=, and still in container's real env as TEST2=
	env3 := "TEST3"     // should not in container's real env
	env4 := "TEST4=a b" // valid
	env5 := "TEST5=a=b" // valid
	res := command.PouchRun("create",
		"--name", name,
		"-e", env1,
		"-e", env2,
		"-e", env3,
		"-e", env4,
		"-e", env5,
		busyboxImage,
		"top")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	envs, err := inspectFilter(name, ".Config.Env")
	c.Assert(err, check.IsNil)

	// check if these envs are in inspect result of container.
	if !strings.Contains(envs, env1) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env1, envs)
	}
	if !strings.Contains(envs, env2) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env2, envs)
	}
	if strings.Contains(envs, env3) {
		c.Fatalf("container env in inspect result should not have %s in %s, while it has\n", env3, envs)
	}
	if !strings.Contains(envs, env4) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env4, envs)
	}
	if !strings.Contains(envs, env5) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env5, envs)
	}

	// check if these envs are in the real container envs
	res = command.PouchRun("start", name)
	res.Assert(c, icmd.Success)

	ret := command.PouchRun("exec", name, "env")
	envs = ret.Stdout()

	if !strings.Contains(envs, env1) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env1, envs)
	}
	if !strings.Contains(envs, env2) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env2, envs)
	}
	//  container's runtime env should not have env3
	if strings.Contains(envs, env3) {
		c.Fatalf("container's runtime env should not have %s in %s while it is there\n", env3, envs)
	}
	if !strings.Contains(envs, env4) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env4, envs)
	}
	if !strings.Contains(envs, env5) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env5, envs)
	}
}

// TestCreateWithEnvfile tests creating container with envfile
func (suite *PouchCreateSuite) TestCreateWithEnvfile(c *check.C) {
	name := "TestCreateWithEnvfile"

	content := "TEST1=value1\n地址1=杭州\n地址2=Hangzhou\naddress=杭州"
	validfile, err := util.TmpFileWithContent(content)
	if err != nil {
		c.Fatal(err)
	}
	defer os.Remove(validfile)

	content = "TEST2="
	invalidfile, err2 := util.TmpFileWithContent(content)
	if err2 != nil {
		c.Fatal(err2)
	}
	defer os.Remove(invalidfile)

	env1 := "TEST1=value1"
	env2 := "TEST2="
	env3 := "TEST3=value3"
	env4 := "地址1=杭州"
	env5 := "地址2=Hangzhou"
	env6 := "address=杭州"
	res := command.PouchRun("create",
		"--name", name,
		"--env-file", validfile,
		"--env-file", invalidfile,
		"-e", env3,
		busyboxImage,
		"top")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	envs, err := inspectFilter(name, ".Config.Env")
	c.Assert(err, check.IsNil)

	// check if these envs are in inspect result of container.
	if !strings.Contains(envs, env1) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env1, envs)
	}
	if !strings.Contains(envs, env2) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env2, envs)
	}
	if !strings.Contains(envs, env3) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env3, envs)
	}
	if !strings.Contains(envs, env4) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env4, envs)
	}
	if !strings.Contains(envs, env5) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env5, envs)
	}
	if !strings.Contains(envs, env6) {
		c.Fatalf("container env in inspect result should have %s in %s while no\n", env6, envs)
	}

	// check if these envs are in the real container envs
	res = command.PouchRun("start", name)
	res.Assert(c, icmd.Success)

	ret := command.PouchRun("exec", name, "env")
	ret.Assert(c, icmd.Success)
	envs = ret.Stdout()

	if !strings.Contains(envs, env1) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env1, envs)
	}
	if !strings.Contains(envs, env2) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env2, envs)
	}
	if !strings.Contains(envs, env3) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env3, envs)
	}
	if !strings.Contains(envs, env4) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env4, envs)
	}
	if !strings.Contains(envs, env5) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env5, envs)
	}
	if !strings.Contains(envs, env6) {
		c.Fatalf("container's runtime env should have %s in %s while no\n", env6, envs)
	}
}

// TestCreateWithWorkDir tests creating container with a workdir works.
// TestCreateWithWorkDir tests creating container with a workdir works.
func (suite *PouchCreateSuite) TestCreateWithWorkDir(c *check.C) {
	name := "TestCreateWithWorkDir"

	res := command.PouchRun("create", "--name", name, "-w /tmp/test", busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	workingDir, err := inspectFilter(name, ".Config.WorkingDir")
	c.Assert(err, check.IsNil)
	c.Assert(strings.TrimSpace(workingDir), check.Equals, "/tmp/test")

	// TODO: check the work directory has been created.
}

// TestCreateWithUser tests creating container with user works.
func (suite *PouchCreateSuite) TestCreateWithUser(c *check.C) {
	name := "TestCreateWithUser"
	user := "1001"

	command.PouchRun("create", "--name", name, "--user", user, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	userConfig, err := inspectFilter(name, ".Config.User")
	c.Assert(err, check.IsNil)
	c.Assert(userConfig, check.Equals, user)
}

// TestCreateWithIntelRdt tests creating container with Intel RDT.
func (suite *PouchCreateSuite) TestCreateWithIntelRdt(c *check.C) {
	name := "TestCreateWithIntelRdt"
	intelRdt := "L3:<cache_id0>=<cbm0>"

	command.PouchRun("create", "--name", name, "--intel-rdt-l3-cbm", intelRdt, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, name)

	intelRdtL3Cbm, err := inspectFilter(name, ".HostConfig.IntelRdtL3Cbm")
	c.Assert(err, check.IsNil)
	c.Assert(intelRdtL3Cbm, check.Equals, intelRdt)
}

// TestCreateWithOOMOption tests creating container with OOM options.
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
		"--annotation", "k1=v1,v2", "--name", cname, busyboxImage)
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
	c.Assert(util.PartialEqual(annotationStr, "k1=v1,v2"), check.IsNil)
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
func (suite *PouchCreateSuite) TestCreateWithPidsLimit(c *check.C) {
	cname := "TestCreateWithPidsLimit"
	command.PouchRun("create", "--pids-limit", "10", "--name", cname, busyboxImage).Assert(c, icmd.Success)

	pidsLimit, err := inspectFilter(cname, ".HostConfig.PidsLimit")
	c.Assert(err, check.IsNil)
	c.Assert(pidsLimit, check.Equals, "10")
}

// TestCreateWithNonExistImage tests running container with image not exist.
func (suite *PouchCreateSuite) TestCreateWithNonExistImage(c *check.C) {
	cname := "TestCreateWithNonExistImage"
	// we should use a non-used image, since containerd not remove image immediately.
	DelImageForceOk(c, busyboxImage)
	command.PouchRun("create", "--name", cname, busyboxImage).Assert(c, icmd.Success)
}

// TestCreateWithNonExistImage tests running container with image not exist.
func (suite *PouchCreateSuite) TestCreateWithNvidiaConfig(c *check.C) {
	cname := "TestCreateWithNvidiaConfig"
	command.PouchRun("create", "--name", cname,
		"--nvidia-capabilities", "all",
		"--nvidia-visible-devs", "none", busyboxImage).Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cname)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	cap := result[0].HostConfig.Resources.NvidiaConfig.NvidiaDriverCapabilities
	drv := result[0].HostConfig.Resources.NvidiaConfig.NvidiaVisibleDevices
	c.Assert(cap, check.Equals, "all")
	c.Assert(drv, check.Equals, "none")
}

// TestCreateWithNonExistImage tests running container with image not exist.
func (suite *PouchCreateSuite) TestCreateWithoutNvidiaConfig(c *check.C) {
	cname := "TestCreateWithoutNvidiaConfig"

	command.PouchRun("create", "--name", cname, busyboxImage).Assert(c, icmd.Success)
	defer command.PouchRun("rm", "-vf", cname)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].HostConfig.Resources.NvidiaConfig, check.IsNil)
}

// TestCreateWithInvalidName tests creating container with invalid name.
func (suite *PouchRunSuite) TestCreateWithInvalidName(c *check.C) {
	name := "new:invalid"
	res := command.PouchRun("create", "--name", name, busyboxImage)
	defer DelContainerForceMultyTime(c, name)
	if !strings.Contains(res.Stdout(), "Invalid container name") {
		check.Commentf("Expected '%s', but got %q", "Invalid container name", res.Stdout())
	}
}
