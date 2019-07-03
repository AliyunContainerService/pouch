package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/daemonv2"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchDaemonSuite is the test suite for daemon.
type PouchDaemonSuite struct{}

func init() {
	check.Suite(&PouchDaemonSuite{})
	os.RemoveAll(daemon.HomeDir)
}

// SetUpTest does common setup in the beginning of each test.
func (suite *PouchDaemonSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)
}

func (suite *PouchDaemonSuite) TearDownTest(c *check.C) {
	os.Remove(daemon.ConfigFile)
}

// TestDaemonCgroupParent tests daemon with cgroup parent
func (suite *PouchDaemonSuite) TestDaemonCgroupParent(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"cgroup-parent": "tmp",
	})

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	cname := "TestDaemonCgroupParent"
	{

		result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}
	{
		result := RunWithSpecifiedDaemon(dcfg, "run",
			"-d", "--name", cname, busyboxImage, "top")
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}
	defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)

	// test if the value is in inspect result
	output := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	//// test if cgroup has the right parent path
	//path := fmt.Sprintf("/sys/fs/cgroup/memory/tmp/%s", result.ID)
	//_, err = os.Stat(path)
	//if err != nil {
	//	daemon.DConfig.DumpLog()
	//	c.Fatalf("get cgroup path failed, err:%s", err)
	//}
}

// TestDaemonConfigFileConflict tests start daemon with configure file conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonConfigFileConflict(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{"containerd-path": "abc"}, "--containerd-path", "def")
	dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonNestObjectConflict tests start daemon with configure file contains nest objects conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonNestObjectConflict(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"TLS": map[string]string{
			"tlscacert": "ca",
			"tlscert":   "cert",
			"tlskey":    "key",
		},
	}, "--tlscacert", "ca")
	dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonSliceFlagNotConflict tests start daemon with configure file contains slice flag will not conflicts with parameter.
func (suite *PouchDaemonSuite) TestDaemonSliceFlagNotConflict(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"label": []string{"a=a", "b=b"},
	}, "--label", "c=d")
	dcfg.KillDaemon()
	c.Assert(err, check.IsNil)
}

// TestDaemonInvalideArgs tests invalid args in daemon return error
func (suite *PouchDaemonSuite) TestDaemonInvalideArgs(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil, "--config=xxx")
	dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonRestart tests daemon restart with running container.
func (suite *PouchDaemonSuite) TestDaemonRestart(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)

	// Must kill it, as we may loose the pid in next call.
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	{
		result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err:%v", result)
		}
	}

	cname := "TestDaemonRestart"
	{
		result := RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname, busyboxImage, "top")
		defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err:%v", result)
		}
	}

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	// test if the container is running.
	output := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(result[0].State.Status), check.Equals, "running")
}

// TestDaemonRestartWithPausedContainer tests daemon with paused container.
func (suite *PouchDaemonSuite) TestDaemonRestartWithPausedContainer(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	{
		result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pull image failed, err: %v", result)
		}
	}

	cname := "TestDaemonRestartWithPausedContainer"
	{
		result := RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname, busyboxImage, "top")
		defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("run container failed, err: %v", result)
		}

		// pause the container
		result = RunWithSpecifiedDaemon(dcfg, "pause", cname)
		if result.ExitCode != 0 {
			dcfg.DumpLog()
			c.Fatalf("pause container failed, err: %v", result)
		}
	}

	// restart daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	// test if the container is paused.
	output := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	data := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(data[0].State.Status), check.Equals, "paused")

	// unpause the container
	result := RunWithSpecifiedDaemon(dcfg, "unpause", cname)
	if result.ExitCode != 0 {
		dcfg.DumpLog()
		c.Fatalf("unpause container failed, err: %v", result)
	}

	//test if the container is running
	output = RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		c.Fatalf("failed to decode inspect output: %v", err)
	}
	c.Assert(string(data[0].State.Status), check.Equals, "running")
}

// TestDaemonLabel tests start daemon with label works.
func (suite *PouchDaemonSuite) TestDaemonLabel(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"label": []string{"a=b"},
	})
	c.Assert(err, check.IsNil)
	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Stdout(), "a=b")
	c.Assert(err, check.IsNil)
}

// TestDaemonLabelDup tests start daemon with duplicated label works.
func (suite *PouchDaemonSuite) TestDaemonLabelDup(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"label": []string{"a=b", "a=b"},
	})
	c.Assert(err, check.IsNil)
	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Stdout(), "a=b")
	c.Assert(err, check.IsNil)

	cnt := strings.Count(result.Stdout(), "a=b")
	c.Assert(cnt, check.Equals, 1)
}

// TestDaemonLabelNeg tests start daemon with wrong label could not work.
func (suite *PouchDaemonSuite) TestDaemonLabelNeg(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"label": []string{"adsf"},
	})
	defer dcfg.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestDaemonDefaultRegistry tests set default registry works.
func (suite *PouchDaemonSuite) TestDaemonDefaultRegistry(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"default-registry":           "registry.hub.docker.com",
		"default-registry-namespace": "library",
	})

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	// Check pull image with default registry using the registry specified in daemon.
	result := RunWithSpecifiedDaemon(dcfg, "pull", "nginx:latest")
	err = util.PartialEqual(result.Combined(), "registry.hub.docker.com/library/nginx:latest")
	c.Assert(err, check.IsNil)
}

// TestDaemonCriEnabled tests enabling cri part in pouchd.
func (suite *PouchDaemonSuite) TestDaemonCriEnabled(c *check.C) {
	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"enable-cri": true,
	})

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	err = util.PartialEqual(result.Combined(), "CriEnabled: true")
	c.Assert(err, check.IsNil)
}

// TestDaemonTlsVerify tests start daemon with TLS verification enabled.
func (suite *PouchDaemonSuite) TestDaemonTlsVerify(c *check.C) {
	SkipIfFalse(c, IsTLSExist)
	dcfg, _ := StartDefaultDaemon(map[string]interface{}{
		"listen": []string{testDaemonHTTPSAddr, daemon.Listen},
		"TLS": map[string]interface{}{
			"tlsverify": true,
			"tlscacert": serverCa,
			"tlscert":   serverCert,
			"tlskey":    serverKey,
		},
	})

	// Must kill it, as we may loose the pid in next call.
	defer dcfg.KillDaemon()

	// Use TLS could success
	result := RunWithSpecifiedDaemon(dcfg,
		"--tlscacert="+clientCa,
		"--tlscert="+clientCert,
		"--tlskey="+clientKey, "version")
	result.Assert(c, icmd.Success)

	// Do not use TLS should fail
	result = RunWithSpecifiedDaemon(dcfg, "version")
	c.Assert(result.ExitCode, check.Equals, 1)
	err := util.PartialEqual(result.Stderr(), "Client sent an HTTP request to an HTTPS server")
	c.Assert(err, check.IsNil)

	{
		// Use wrong CA should fail
		result := RunWithSpecifiedDaemon(dcfg,
			"--tlscacert="+clientWrongCa,
			"--tlscert="+clientCert,
			"--tlskey="+clientKey, "version")
		c.Assert(result.ExitCode, check.Equals, 1)
		err := util.PartialEqual(result.Stderr(), "failed to append certificates")
		c.Assert(err, check.IsNil)
	}
}

// TestDaemonStartOverOneTimes tests start daemon over one times should fail.
func (suite *PouchDaemonSuite) TestDaemonStartOverOneTimes(c *check.C) {
	addr1 := "unix:///var/run/pouchtest1.sock"
	dcfg1, err := StartDefaultDaemon(map[string]interface{}{
		"listen": []string{addr1},
	})
	c.Assert(err, check.IsNil)
	defer dcfg1.KillDaemon()

	// verify listen to tcp works
	command.PouchRun("--host", addr1, "version").Assert(c, icmd.Success)

	// test second daemon with same pidfile should start fail
	addr2 := "unix:///var/run/pouchtest2.sock"
	dcfg2, err := StartDefaultDaemon(map[string]interface{}{
		"listen": []string{addr2},
	})

	defer dcfg2.KillDaemon()
	c.Assert(err, check.NotNil)
}

// TestRestartStoppedContainerAfterDaemonRestart is used to test the case that
// when container is stopped and then pouchd restarts, the restore logic should
// initialize the existing container IO settings even though they are not alive.
func (suite *PouchDaemonSuite) TestRestartStoppedContainerAfterDaemonRestart(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	var (
		cname = c.TestName()
		msg   = "hello"
	)

	// pull image
	RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage).Assert(c, icmd.Success)

	// run a container
	res := RunWithSpecifiedDaemon(dcfg, "run", "--name", cname, busyboxImage, "echo", msg)
	defer ensureContainerNotExist(dcfg, cname)

	res.Assert(c, icmd.Success)
	c.Assert(strings.TrimSpace(res.Combined()), check.Equals, msg)

	// wait for it.
	RunWithSpecifiedDaemon(dcfg, "wait", cname).Assert(c, icmd.Success)

	// waiting for containerd meta data cleaning.
	time.Sleep(time.Second * 5)

	// kill the daemon and make sure it has been killed
	dcfg.KillDaemon()
	c.Assert(dcfg.IsDaemonUp(), check.Equals, false)

	// restart again
	c.Assert(dcfg.StartDaemon(), check.IsNil)
	defer dcfg.KillDaemon()

	// start the container again
	res = RunWithSpecifiedDaemon(dcfg, "start", "-a", cname)
	res.Assert(c, icmd.Success)
	c.Assert(strings.TrimSpace(res.Combined()), check.Equals, msg)
}

// TestUpdateDaemonWithLabels tests update daemon online with labels updated
func (suite *PouchDaemonSuite) TestUpdateDaemonWithLabels(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	RunWithSpecifiedDaemon(dcfg, "updatedaemon", "--label", "aaa=bbb").Assert(c, icmd.Success)

	ret := RunWithSpecifiedDaemon(dcfg, "info")
	ret.Assert(c, icmd.Success)

	updated := strings.Contains(ret.Stdout(), "aaa=bbb")
	c.Assert(updated, check.Equals, true)
}

// TestUpdateDaemonOffline tests update daemon offline
func (suite *PouchDaemonSuite) TestUpdateDaemonOffline(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	RunWithSpecifiedDaemon(dcfg, "updatedaemon", "--offline=true").Assert(c, icmd.Success)

	ret := RunWithSpecifiedDaemon(dcfg, "info")
	ret.Assert(c, icmd.Success)
}

func ensureContainerNotExist(dcfg *daemon.Config, cname string) error {
	_ = RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)
	return nil
}

// TestRecoverContainerWhenHostDown tests when the host down, the pouchd still can
// recover the container whose restart policy is always .
func (suite *PouchDaemonSuite) TestRecoverContainerWhenHostDown(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)
	//Start a test daemon with test args.
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	cname := "TestRecoverContainerWhenHostDown"
	ensureContainerNotExist(dcfg, cname)

	// prepare test image
	result := RunWithSpecifiedDaemon(dcfg, "pull", busyboxImage)
	if result.ExitCode != 0 {
		dcfg.DumpLog()
		c.Fatalf("pull image failed, err: %v", result)
	}

	result = RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname, "--restart", "always", busyboxImage, "top")
	if result.ExitCode != 0 {
		dcfg.DumpLog()
		c.Fatalf("run container failed, err: %v", result)
	}
	defer ensureContainerNotExist(dcfg, cname)

	// get the container init process id
	pidStr := RunWithSpecifiedDaemon(dcfg, "inspect", "-f", "{{.State.Pid}}", cname).Stdout()

	// get parent pid of container init process
	output, err := exec.Command("ps", "-o", "ppid=", "-p", strings.TrimSpace(pidStr)).Output()
	if err != nil {
		c.Errorf("failed to get parent pid of container %s: output: %s err: %v", cname, string(output), err)
	}
	// imitate the host down
	// first kill the daemon
	dcfg.KillDaemon()

	// second kill the container's process
	ppid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		dcfg.DumpLog()
		c.Fatalf("failed to convert pid string %s to int: %v", output, err)
	}
	syscall.Kill(ppid, syscall.SIGKILL)

	// restart the daemon
	err = RestartDaemon(dcfg)
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	// wait container started again or timeout error
	check := make(chan struct{})
	timeout := make(chan bool, 1)
	// set timeout to wait container started
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	// check whether container started
	go func() {
		for {
			data := RunWithSpecifiedDaemon(dcfg, "inspect", cname).Stdout()
			cInfo := []types.ContainerJSON{}
			if err := json.Unmarshal([]byte(data), &cInfo); err != nil {
				c.Fatalf("failed to decode inspect output: %v", err)
			}

			if len(cInfo) == 0 || cInfo[0].State == nil {
				continue
			}

			if string(cInfo[0].State.Status) == "running" {
				check <- struct{}{}
				break
			}

			fmt.Printf("container %s status: %s\n", cInfo[0].ID, string(cInfo[0].State.Status))
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-check:
	case <-timeout:
		dcfg.DumpLog()
		c.Fatalf("failed to wait container running")
	}
}

// TestDaemonWithSysyemdCgroupDriver tests start daemon with systemd cgroup driver
func (suite *PouchDaemonSuite) TestDaemonWithSystemdCgroupDriver(c *check.C) {
	SkipIfFalse(c, environment.SupportSystemdCgroupDriver)

	dcfg, err := StartDefaultDaemon(map[string]interface{}{
		"cgroup-driver": "systemd",
	})
	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	result := RunWithSpecifiedDaemon(dcfg, "info")
	c.Assert(util.PartialEqual(result.Stdout(), "systemd"), check.IsNil)

	cname := "TestWithSystemdCgroupDriver"
	ret := RunWithSpecifiedDaemon(dcfg, "run", "-d", "--name", cname, busyboxImage, "top")
	defer RunWithSpecifiedDaemon(dcfg, "rm", "-f", cname)
	ret.Assert(c, icmd.Success)
}

// TestContainerdPIDReuse tests even though old containerd pid being reused, we can still
// pull up the containerd instance.
func (suite *PouchDaemonSuite) TestContainerdPIDReuse(c *check.C) {
	containerStateDir := filepath.Join(daemon.HomeDir, "containerd/state")
	err := os.MkdirAll(containerStateDir, 0664)
	c.Assert(err, check.IsNil)

	containerdPidPath := filepath.Join(containerStateDir, "containerd.pid")

	// set containerd pid to 1 to make sure the pid must be alive
	if err := ioutil.WriteFile(containerdPidPath, []byte(fmt.Sprintf("%d", 1)), 0660); err != nil {
		c.Errorf("failed to write pid to file: %v", containerdPidPath)
	}

	// make sure pouchd can successfully start
	dcfg, err := StartDefaultDaemon(nil)

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()
}

// TestUpdateDaemonWithHomeDirAndSnapshotter tests update daemon with home-dir and snapshotter
func (suite *PouchDaemonSuite) TestUpdateDaemonWithHomeDirAndSnapshotter(c *check.C) {
	dcfg, err := StartDefaultDaemon(nil)

	c.Assert(err, check.IsNil)
	defer dcfg.KillDaemon()

	tmpHomeDir := "/tmp/pouch_dir"
	snapshotter := "test_snapshotter"

	RunWithSpecifiedDaemon(dcfg, "updatedaemon", "--config-file", daemon.ConfigFile, "--offline=true", "--home-dir", tmpHomeDir, "--snapshotter", snapshotter).Assert(c, icmd.Success)

	ret := RunWithSpecifiedDaemon(dcfg, "info")
	ret.Assert(c, icmd.Success)

	f, err := os.Open(daemon.ConfigFile)
	c.Assert(err, check.IsNil)
	defer f.Close()

	readConfig := config.Config{}

	err = json.NewDecoder(f).Decode(&readConfig)
	c.Assert(err, check.IsNil)

	c.Assert(readConfig.HomeDir, check.Equals, tmpHomeDir)
	c.Assert(readConfig.Snapshotter, check.Equals, snapshotter)
}

// TestUpdateDaemonWithDisableBridge tests update daemon with disable bridge network
func (suite *PouchDaemonSuite) TestUpdateDaemonWithDisableBridge(c *check.C) {
	d := daemonv2.New()

	// modify test config
	d.Config.NetworkConfig.BridgeConfig.DisableBridge = true

	err := d.Start()
	if err != nil {
		c.Fatalf("failed to start daemon with json, err(%v)", err)
	}
	defer d.Clean()

	res := d.RunCommand("network", "ls")
	res.Assert(c, icmd.Success)

	if strings.Contains(res.Stdout(), "bridge") {
		d.RunCommand("network", "rm", "bridge").Assert(c, icmd.Success)
	}

	d.Restart()

	res = d.RunCommand("network", "ls")
	res.Assert(c, icmd.Success)
	if strings.Contains(res.Stdout(), "bridge") {
		c.Fatalf("failed to disable bridge network")
	}
}
