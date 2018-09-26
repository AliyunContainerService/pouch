package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
)

const (
	cpTestPathParent = "/some"
	cpTestPath       = "/some/path"
	cpTestName       = "test"
	cpFullPath       = "/some/path/test"

	cpContainerContents = "holla, i am the container"
	cpHostContents      = "hello, i am the host"
)

// PouchCpSuite is the test suite for cp CLI.
type PouchCpSuite struct{}

func init() {
	check.Suite(&PouchCpSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchCpSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

func runPouchCp(c *check.C, src, dst string) (err error) {
	c.Logf("running `docker cp %s %s`", src, dst)

	args := []string{"cp", src, dst}

	out, _, err := RunCommandWithOutput(exec.Command("pouch", args...))
	if err != nil {
		err = fmt.Errorf("error executing `pouch cp` command: %s: %s", err, out)
	}

	return
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchCpSuite) TearDownTest(c *check.C) {
}

// Ensure that an all-local path case returns an error.
func (suite *PouchCpSuite) TestCpLocalOnly(c *check.C) {
	err := runPouchCp(c, "foo", "bar")
	c.Assert(err, check.NotNil)

	c.Assert(strings.Contains(err.Error(), "must specify at least one container source"), check.Equals, true)
}

// Check that garbage paths don't escape the container's rootfs
func (suite *PouchCpSuite) TestCpGarbagePath(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	c.Assert(os.MkdirAll(cpTestPath, os.ModeDir), check.IsNil)

	hostFile, err := os.Create(cpFullPath)
	c.Assert(err, check.IsNil)
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "pouch-integration")
	c.Assert(err, check.IsNil)

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := path.Join("../../../../../../../../../../../../", cpFullPath)

	command.PouchRun("cp", containerID+":"+path, tmpdir)

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)

	// output matched host file -- garbage path can escape container rootfs
	c.Assert(string(test), check.Not(check.Equals), cpHostContents)

	// output doesn't match the input for garbage path
	c.Assert(string(test), check.Equals, cpContainerContents)

	command.PouchRun("rm", "-f", containerID)
}

// Check that relative paths are relative to the container's rootfs
func (suite *PouchCpSuite) TestCpRelativePath(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	c.Assert(os.MkdirAll(cpTestPath, os.ModeDir), check.IsNil)

	hostFile, err := os.Create(cpFullPath)
	c.Assert(err, check.IsNil)
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "pouch-integration")
	c.Assert(err, check.IsNil)

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	var relPath string
	if path.IsAbs(cpFullPath) {
		// normally this is `filepath.Rel("/", cpFullPath)` but we cannot
		// get this unix-path manipulation on windows with filepath.
		relPath = cpFullPath[1:]
	}
	c.Assert(path.IsAbs(cpFullPath), check.Equals, true)

	command.PouchRun("cp", containerID+":"+relPath, tmpdir)

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)

	// output matched host file -- relative path can escape container rootfs
	c.Assert(string(test), check.Not(check.Equals), cpHostContents)

	// output doesn't match the input for relative path
	c.Assert(string(test), check.Equals, cpContainerContents)

	command.PouchRun("rm", "-f", containerID)
}

// Check that absolute paths are relative to the container's rootfs
func (suite *PouchCpSuite) TestCpAbsolutePath(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	c.Assert(os.MkdirAll(cpTestPath, os.ModeDir), check.IsNil)

	hostFile, err := os.Create(cpFullPath)
	c.Assert(err, check.IsNil)
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "pouchd-integration")
	c.Assert(err, check.IsNil)

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := cpFullPath

	command.PouchRun("cp", containerID+":"+path, tmpdir)

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)

	// output matched host file -- absolute path can escape container rootfs
	c.Assert(string(test), check.Not(check.Equals), cpHostContents)

	// output doesn't match the input for absolute path
	c.Assert(string(test), check.Equals, cpContainerContents)

	out = command.PouchRun("rm", "-f", containerID)
	// failed to set up container
	c.Assert(strings.TrimSpace(out.Stdout()), check.Equals, containerID)
}

// Check that absolute symlinks are still relative to the container's rootfs
func (suite *PouchCpSuite) TestCpAbsoluteSymlink(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpFullPath+" container_path && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	c.Assert(os.MkdirAll(cpTestPath, os.ModeDir), check.IsNil)

	hostFile, err := os.Create(cpFullPath)
	c.Assert(err, check.IsNil)
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	c.Assert(err, check.IsNil)

	tmpname := filepath.Join(tmpdir, "container_path")
	defer os.RemoveAll(tmpdir)

	path := path.Join("/", "container_path")

	command.PouchRun("cp", containerID+":"+path, tmpdir)

	// We should have copied a symlink *NOT* the file itself!
	linkTarget, err := os.Readlink(tmpname)
	c.Assert(err, check.IsNil)

	c.Assert(linkTarget, check.Equals, filepath.FromSlash(cpFullPath))

	command.PouchRun("rm", "-f", containerID)
}

// Check that symlinks to a directory behave as expected when copying one from
// a container.
func (suite *PouchCpSuite) TestCpFromSymlinkToDirectory(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpTestPathParent+" /dir_link && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	testDir, err := ioutil.TempDir("", "test-cp-from-symlink-to-dir-")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(testDir)

	// This copy command should copy the symlink, not the target, into the
	// temporary directory.
	command.PouchRun("cp", containerID+":"+"/dir_link", testDir)

	expectedPath := filepath.Join(testDir, "dir_link")
	linkTarget, err := os.Readlink(expectedPath)
	c.Assert(err, check.IsNil)

	c.Assert(linkTarget, check.Equals, filepath.FromSlash(cpTestPathParent))

	os.Remove(expectedPath)

	// This copy command should resolve the symlink (note the trailing
	// separator), copying the target into the temporary directory.
	command.PouchRun("cp", containerID+":"+"/dir_link/", testDir)

	// It *should not* have copied the directory using the target's name, but
	// used the given name instead.
	unexpectedPath := filepath.Join(testDir, cpTestPathParent)
	stat, err := os.Lstat(unexpectedPath)

	var outInfo string
	if err == nil {
		outInfo = fmt.Sprintf("target name was copied: %q - %q", stat.Mode(), stat.Name())
	}
	c.Assert(err, check.NotNil, check.Commentf(outInfo))

	// It *should* have copied the directory using the asked name "dir_link".
	stat, err = os.Lstat(expectedPath)
	c.Assert(err, check.IsNil, check.Commentf("unable to stat resource at %q", expectedPath))

	c.Assert(stat.IsDir(), check.Equals, true)

	command.PouchRun("rm", "-f", containerID)
}

// Check that symlinks which are part of the resource path are still relative to the container's rootfs
func (suite *PouchCpSuite) TestCpSymlinkComponent(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpTestPath+" container_path && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	c.Assert(os.MkdirAll(cpTestPath, os.ModeDir), check.IsNil)

	hostFile, err := os.Create(cpFullPath)
	c.Assert(err, check.IsNil)
	defer hostFile.Close()
	defer os.RemoveAll(cpTestPathParent)

	fmt.Fprintf(hostFile, "%s", cpHostContents)

	tmpdir, err := ioutil.TempDir("", "pouch-integration")

	c.Assert(err, check.IsNil)

	tmpname := filepath.Join(tmpdir, cpTestName)
	defer os.RemoveAll(tmpdir)

	path := path.Join("/", "container_path", cpTestName)

	command.PouchRun("cp", containerID+":"+path, tmpdir)

	file, _ := os.Open(tmpname)
	defer file.Close()

	test, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)

	// output matched host file -- symlink path component can escape container rootfs
	c.Assert(string(test), check.Not(check.Equals), cpHostContents)

	// output doesn't match the input for symlink path component
	c.Assert(string(test), check.Equals, cpContainerContents)

	command.PouchRun("rm", "-f", containerID)
}

func (suite *PouchCpSuite) TestCpSpecialFiles(c *check.C) {
	outDir, err := ioutil.TempDir("", "cp-test-special-files")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(outDir)

	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "touch /foo  && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	// Copy actual /etc/resolv.conf
	command.PouchRun("cp", containerID+":/etc/resolv.conf", outDir)

	expected, err := readContainerFile(containerID, "/etc/resolv.conf")
	c.Assert(err, check.IsNil)
	actual, err := ioutil.ReadFile(outDir + "/resolv.conf")
	c.Assert(err, check.IsNil)

	// Expected copied file to be duplicate of the container resolvconf
	c.Assert(bytes.Equal(actual, expected), check.Equals, true)

	// Copy actual /etc/hosts
	command.PouchRun("cp", containerID+":/etc/hosts", outDir)

	expected, err = readContainerFile(containerID, "/etc/hosts")
	c.Assert(err, check.IsNil)
	actual, err = ioutil.ReadFile(outDir + "/hosts")
	c.Assert(err, check.IsNil)

	// Expected copied file to be duplicate of the container hosts
	c.Assert(bytes.Equal(actual, expected), check.Equals, true)

	// Copy actual /etc/resolv.conf
	command.PouchRun("cp", containerID+":/etc/hostname", outDir)

	expected, err = readContainerFile(containerID, "/etc/hostname")
	c.Assert(err, check.IsNil)
	actual, err = ioutil.ReadFile(outDir + "/hostname")
	c.Assert(err, check.IsNil)

	// Expected copied file to be duplicate of the container resolvconf
	c.Assert(bytes.Equal(actual, expected), check.Equals, true)

	command.PouchRun("rm", "-f", containerID)
}

func (suite *PouchCpSuite) TestCpToDot(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "echo lololol > /test  && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	tmpdir, err := ioutil.TempDir("", "docker-integration")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(tmpdir)
	cwd, err := os.Getwd()
	c.Assert(err, check.IsNil)
	defer os.Chdir(cwd)
	c.Assert(os.Chdir(tmpdir), check.IsNil)
	command.PouchRun("cp", containerID+":/test", ".")
	content, err := ioutil.ReadFile("./test")
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "lololol\n")
}

func (suite *PouchCpSuite) TestCpToStdout(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "echo lololol > /test  && sleep 100")

	containerID := strings.TrimSpace(out.Stdout())

	outInfo, _, err := RunCommandPipelineWithOutput(
		exec.Command("pouch", "cp", containerID+":/test", "-"),
		exec.Command("tar", "-vtf", "-"))

	c.Assert(err, check.IsNil)

	c.Assert(strings.Contains(outInfo, "test"), check.Equals, true)
	c.Assert(strings.Contains(outInfo, "-rw"), check.Equals, true)
}

func (suite *PouchCpSuite) TestCopyCreatedContainer(c *check.C) {
	command.PouchRun("create", "--name", "test_cp", "-v", "/test", "registry.hub.docker.com/library/busybox:1.28")

	tmpDir, err := ioutil.TempDir("", "test")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(tmpDir)
	command.PouchRun("cp", "test_cp:/bin/sh", tmpDir)
}

// test copy with option `-L`: following symbol link
// Check that symlinks to a file behave as expected when copying one from
// a container to host following symbol link
func (suite *PouchCpSuite) TestCpSymlinkFromConToHostFollowSymlink(c *check.C) {
	out := command.PouchRun("run", "-d", "registry.hub.docker.com/library/busybox:1.28", "/bin/sh", "-c", "mkdir -p '"+cpTestPath+"' && echo -n '"+cpContainerContents+"' > "+cpFullPath+" && ln -s "+cpFullPath+" /dir_link && sleep 100")
	if out.ExitCode != 0 {
		c.Fatal("failed to create a container", out)
	}

	cleanedContainerID := strings.TrimSpace(out.Stdout())

	testDir, err := ioutil.TempDir("", "test-cp-symlink-container-to-host-follow-symlink")
	if err != nil {
		c.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// This copy command should copy the symlink, not the target, into the
	// temporary directory.
	command.PouchRun("cp", "-L", cleanedContainerID+":"+"/dir_link", testDir)

	expectedPath := filepath.Join(testDir, "dir_link")

	expected := []byte(cpContainerContents)
	actual, err := ioutil.ReadFile(expectedPath)
	c.Assert(err, check.IsNil)

	if !bytes.Equal(actual, expected) {
		c.Fatalf("Expected copied file to be duplicate of the container symbol link target")
	}
	os.Remove(expectedPath)

	// now test copy symbol link to a non-existing file in host
	expectedPath = filepath.Join(testDir, "somefile_host")
	// expectedPath shouldn't exist, if exists, remove it
	if _, err := os.Lstat(expectedPath); err == nil {
		os.Remove(expectedPath)
	}

	command.PouchRun("cp", "-L", cleanedContainerID+":"+"/dir_link", expectedPath)

	actual, err = ioutil.ReadFile(expectedPath)
	c.Assert(err, check.IsNil)

	if !bytes.Equal(actual, expected) {
		c.Fatalf("Expected copied file to be duplicate of the container symbol link target")
	}
	defer os.Remove(expectedPath)
}
