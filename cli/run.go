package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alibaba/pouch/pkg/reference"

	"github.com/spf13/cobra"
)

// runDescription is used to describe run command in detail and auto generate command doc.
var runDescription = "Create a container object in Pouchd, and start the container. " +
	"This is useful when you just want to use one command to start a container. "

// RunCommand use to implement 'run' command, it creates and starts a container.
type RunCommand struct {
	baseCommand
	container
	detachKeys string
	attach     bool
	stdin      bool
	detach     bool
}

// Init initialize run command.
func (rc *RunCommand) Init(c *Cli) {
	rc.cli = c
	rc.cmd = &cobra.Command{
		Use:   "run [OPTIONS] IMAGE [ARG...]",
		Short: "Create a new container and start it",
		Long:  runDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runRun(args)
		},
		Example: runExample(),
	}
	rc.addFlags()
}

// addFlags adds flags for specific command.
func (rc *RunCommand) addFlags() {
	flagSet := rc.cmd.Flags()
	flagSet.SetInterspersed(false)
	flagSet.StringVar(&rc.name, "name", "", "Specify name of container")
	flagSet.BoolVarP(&rc.tty, "tty", "t", false, "Allocate a tty device")
	flagSet.StringSliceVarP(&rc.volume, "volume", "v", nil, "Bind mount volumes to container")
	flagSet.StringVar(&rc.runtime, "runtime", "", "Specify oci runtime")
	flagSet.StringVar(&rc.detachKeys, "detach-keys", "", "Override the key sequence for detaching a container")
	flagSet.BoolVarP(&rc.attach, "attach", "a", false, "Attach container's STDOUT and STDERR")
	flagSet.BoolVarP(&rc.stdin, "interactive", "i", false, "Attach container's STDIN")
	flagSet.StringSliceVarP(&rc.env, "env", "e", nil, "Set environment variables for container")
	flagSet.StringSliceVarP(&rc.labels, "label", "l", nil, "Set label for a container")
	flagSet.StringVar(&rc.entrypoint, "entrypoint", "", "Overwrite the default entrypoint")
	flagSet.StringVarP(&rc.workdir, "workdir", "w", "", "Set the working directory in a container")
	flagSet.BoolVarP(&rc.detach, "detach", "d", false, "Run container in background and print container ID")
	flagSet.StringVar(&rc.hostname, "hostname", "", "Set container's hostname")
	flagSet.Int64Var(&rc.cpushare, "cpu-share", 0, "CPU shares")
	flagSet.StringVar(&rc.cpusetcpus, "cpuset-cpus", "", "CPUs in cpuset")
	flagSet.StringVar(&rc.cpusetmems, "cpuset-mems", "", "MEMs in cpuset")
	flagSet.Int64Var(&rc.memorySwappiness, "memory-wappiness", -1, "Container memory swappiness [0, 100]")
	flagSet.StringVarP(&rc.memory, "memory", "m", "", "Container memory limit")
	flagSet.StringVar(&rc.memorySwap, "memory-swap", "", "Container swap limit")
	flagSet.StringSliceVarP(&rc.devices, "device", "", nil, "Add a host device to the container")
	flagSet.BoolVar(&rc.enableLxcfs, "enableLxcfs", false, "Enable lxcfs")
	flagSet.StringVar(&rc.restartPolicy, "restart", "", "Restart policy to apply when container exits")
}

// runRun is the entry of run command.
func (rc *RunCommand) runRun(args []string) error {
	config, err := rc.config()
	if err != nil {
		return fmt.Errorf("failed to run container: %v", err)
	}

	ref, err := reference.Parse(args[0])
	if err != nil {
		return fmt.Errorf("failed to run container: %v", err)
	}

	config.Image = ref.String()
	if len(args) > 1 {
		config.Cmd = args[1:]
	}
	containerName := rc.name

	apiClient := rc.cli.Client()
	result, err := apiClient.ContainerCreate(config.ContainerConfig, config.HostConfig, containerName)
	if err != nil {
		return fmt.Errorf("failed to run container: %v", err)
	}
	if len(result.Warnings) != 0 {
		fmt.Printf("WARNING: %s \n", strings.Join(result.Warnings, "\n"))
	}

	// pouch run not specify --name
	if containerName == "" {
		containerName = result.Name
	}

	if (rc.attach || rc.stdin) && rc.detach {
		return fmt.Errorf("Conflicting options: -a (or -i) and -d")
	}

	// default attach container's stdout and stderr
	if !rc.detach {
		rc.attach = true
	}

	wait := make(chan struct{})

	if rc.attach || rc.stdin {
		in, out, err := setRawMode(rc.stdin, false)
		if err != nil {
			return fmt.Errorf("failed to set raw mode")
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		conn, br, err := apiClient.ContainerAttach(containerName, rc.stdin)
		if err != nil {
			return fmt.Errorf("failed to attach container: %v", err)
		}

		go func() {
			io.Copy(os.Stdout, br)
			wait <- struct{}{}
		}()
		go func() {
			io.Copy(conn, os.Stdin)
			wait <- struct{}{}
		}()
	}

	// start container
	if err := apiClient.ContainerStart(containerName, rc.detachKeys); err != nil {
		return fmt.Errorf("failed to run container %s: %v", containerName, err)
	}

	// wait the io to finish
	if rc.attach || rc.stdin {
		<-wait
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", result.ID)
	}
	return nil
}

// runExample shows examples in run command, and is used in auto-generated cli docs.
func runExample() string {
	return `$ pouch run --name test registry.hub.docker.com/library/busybox:latest echo "hi"
hi
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   23f852   stopped   registry.hub.docker.com/library/busybox:latest   runc      4 seconds ago
$ pouch run -d --name test registry.hub.docker.com/library/busybox:latest
90719b5f9a455b3314a49e72e3ecb9962f215e0f90153aa8911882acf2ba2c84
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   90719b   stopped   registry.hub.docker.com/library/busybox:latest   runc      5 seconds ago
$ pouch run --device /dev/zero:/dev/testDev:rwm --name test registry.hub.docker.com/library/busybox:latest ls -l /dev/testDev
crw-rw-rw-    1 root     root        1,   3 Jan  8 09:40 /dev/testnull
	`
}
