package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/request"

	"github.com/gotestyourself/gotestyourself/icmd"
)

var (
	// DefaultRootDir defines the default root dir for pouchd.
	DefaultRootDir string
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath string
)

func init() {
	DefaultRootDir, _ = GetRootDir()
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath = DefaultRootDir + "/volume"
}

// GetRootDir assign the root dir
func GetRootDir() (string, error) {
	resp, err := request.Get("/info")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	got := types.SystemInfo{}
	err = json.NewDecoder(resp.Body).Decode(&got)
	if err != nil {
		return "", err
	}
	return got.PouchRootDir, nil
}

// StartDefaultDaemonDebug starts a daemon with default configuration and debug on.
func StartDefaultDaemonDebug(args ...string) (*daemon.Config, error) {
	cfg := daemon.NewConfig()
	cfg.Debug = true

	cfg.NewArgs(args...)

	return &cfg, cfg.StartDaemon()
}

// StartDefaultDaemon starts a daemon with all default configuration and debug off.
func StartDefaultDaemon(args ...string) (*daemon.Config, error) {
	cfg := daemon.NewConfig()
	cfg.Debug = false

	cfg.NewArgs(args...)

	return &cfg, cfg.StartDaemon()
}

// RestartDaemon restart daemon
func RestartDaemon(cfg *daemon.Config) error {
	cfg.KillDaemon()
	return cfg.StartDaemon()
}

// RunWithSpecifiedDaemon run pouch command with --host parameter
func RunWithSpecifiedDaemon(d *daemon.Config, cmd ...string) *icmd.Result {
	var sock string

	// Find the first -l or --listen parameter and use it.
	for _, v := range d.Args {
		if strings.Contains(v, "-l") || strings.Contains(v, "--listen") {
			if strings.Contains(v, "--listen-cri") {
				continue
			}
			if strings.Contains(v, "=") {
				sock = strings.Split(v, "=")[1]
			} else {
				sock = strings.Fields(v)[1]
			}
			break
		}
	}
	args := append(append([]string{"--host"}, sock), cmd...)
	return command.PouchRun(args...)
}

// CreateConfigFile create configuration file and marshal cfg.
func CreateConfigFile(path string, cfg interface{}) error {
	idx := strings.LastIndex(path, "/")
	if _, err := os.Stat(path[0:idx]); os.IsNotExist(err) {
		os.Mkdir(path[0:idx], os.ModePerm)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	s, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "%s", s)
	file.Sync()

	defer file.Close()
	return nil
}

// RunCommandWithOutput runs the specified command and returns the combined output (stdout/stderr)
// with the exitCode different from 0 and the error if something bad happened
func RunCommandWithOutput(cmd *exec.Cmd) (output string, exitCode int, err error) {
	exitCode = 0
	out, err := cmd.CombinedOutput()
	exitCode = ProcessExitCode(err)
	output = string(out)
	return
}

// ProcessExitCode process the specified error and returns the exit status code
// if the error was of type exec.ExitError, returns nothing otherwise.
func ProcessExitCode(err error) (exitCode int) {
	if err != nil {
		var exiterr error
		if exitCode, exiterr = GetExitCode(err); exiterr != nil {
			// TODO: Fix this so we check the error's text.
			// we've failed to retrieve exit code, so we set it to 127
			exitCode = 127
		}
	}
	return
}

// GetExitCode returns the ExitStatus of the specified error if its type is
// exec.ExitError, returns 0 and an error otherwise.
func GetExitCode(err error) (int, error) {
	exitCode := 0
	if exiterr, ok := err.(*exec.ExitError); ok {
		if procExit, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return procExit.ExitStatus(), nil
		}
	}
	return exitCode, fmt.Errorf("failed to get exit code")
}

// RunCommandPipelineWithOutput runs the array of commands with the output
// of each pipelined with the following (like cmd1 | cmd2 | cmd3 would do).
// It returns the final output, the exitCode different from 0 and the error
// if something bad happened.
func RunCommandPipelineWithOutput(cmds ...*exec.Cmd) (output string, exitCode int, err error) {
	if len(cmds) < 2 {
		return "", 0, fmt.Errorf("pipeline does not have multiple cmds")
	}

	// connect stdin of each cmd to stdout pipe of previous cmd
	for i, cmd := range cmds {
		if i > 0 {
			prevCmd := cmds[i-1]
			cmd.Stdin, err = prevCmd.StdoutPipe()

			if err != nil {
				return "", 0, fmt.Errorf("cannot set stdout pipe for %s: %v", cmd.Path, err)
			}
		}
	}

	// start all cmds except the last
	for _, cmd := range cmds[:len(cmds)-1] {
		if err = cmd.Start(); err != nil {
			return "", 0, fmt.Errorf("starting %s failed with error: %v", cmd.Path, err)
		}
	}

	var pipelineError error
	defer func() {
		// wait all cmds except the last to release their resources
		for _, cmd := range cmds[:len(cmds)-1] {
			if err := cmd.Wait(); err != nil {
				pipelineError = fmt.Errorf("command %s failed with error: %v", cmd.Path, err)
				break
			}
		}
	}()
	if pipelineError != nil {
		return "", 0, pipelineError
	}

	// wait on last cmd
	return RunCommandWithOutput(cmds[len(cmds)-1])
}

func readContainerFile(containerID, filename string) ([]byte, error) {
	rootPath := "/var/lib/pouch/containerd/state/io.containerd.runtime.v1.linux/default/"
	f, err := os.Open(filepath.Join(rootPath, containerID, "rootfs", filename))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return content, nil
}
