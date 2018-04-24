package main

import (
	"context"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

// updateDescription is used to describe update command in detail and auto generate command doc.
var updateDescription = "Update a container's configurations, including memory, cpu and diskquota etc.  " +
	"You can update a container when it is running."

// UpdateCommand use to implement 'update' command, it modifies the configurations of a container.
type UpdateCommand struct {
	baseCommand
	container
	image string
}

// Init initialize update command.
func (uc *UpdateCommand) Init(c *Cli) {
	uc.cli = c
	uc.cmd = &cobra.Command{
		Use:   "update [OPTIONS] CONTAINER",
		Short: "Update the configurations of a container",
		Long:  updateDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return uc.updateRun(args)
		},
		Example: updateExample(),
	}
	uc.addFlags()
}

// addFlags adds flags for specific command.
func (uc *UpdateCommand) addFlags() {
	flagSet := uc.cmd.Flags()
	flagSet.SetInterspersed(false)
	flagSet.Uint16Var(&uc.blkioWeight, "blkio-weight", 0, "Block IO (relative weight), between 10 and 1000, or 0 to disable")
	flagSet.Int64Var(&uc.cpushare, "cpu-share", 0, "CPU shares (relative weight)")
	flagSet.StringVar(&uc.cpusetcpus, "cpuset-cpus", "", "CPUs in cpuset")
	flagSet.StringVar(&uc.cpusetmems, "cpuset-mems", "", "MEMs in cpuset")
	flagSet.StringVarP(&uc.memory, "memory", "m", "", "Container memory limit")
	flagSet.StringVar(&uc.memorySwap, "memory-swap", "", "Container swap limit")
	flagSet.Int64Var(&uc.memorySwappiness, "memory-swappiness", -1, "Container memory swappiness [0, 100]")
	flagSet.StringSliceVarP(&uc.env, "env", "e", nil, "Set environment variables for container")
	flagSet.StringSliceVarP(&uc.labels, "label", "l", nil, "Set label for container")
	flagSet.StringVar(&uc.restartPolicy, "restart", "", "Restart policy to apply when container exits")
}

// updateRun is the entry of update command.
func (uc *UpdateCommand) updateRun(args []string) error {
	container := args[0]
	ctx := context.Background()

	labels, err := opts.ParseLabels(uc.labels)
	if err != nil {
		return err
	}

	if err := opts.ValidateMemorySwappiness(uc.memorySwappiness); err != nil {
		return err
	}

	memory, err := opts.ParseMemory(uc.memory)
	if err != nil {
		return err
	}

	memorySwap, err := opts.ParseMemorySwap(uc.memorySwap)
	if err != nil {
		return err
	}

	resource := types.Resources{
		CPUShares:        uc.cpushare,
		CpusetCpus:       uc.cpusetcpus,
		CpusetMems:       uc.cpusetmems,
		Memory:           memory,
		MemorySwap:       memorySwap,
		MemorySwappiness: &uc.memorySwappiness,
		BlkioWeight:      uc.blkioWeight,
	}

	restartPolicy, err := opts.ParseRestartPolicy(uc.restartPolicy)
	if err != nil {
		return err
	}

	updateConfig := &types.UpdateConfig{
		Env:           uc.env,
		Labels:        labels,
		RestartPolicy: restartPolicy,
		Resources:     resource,
	}

	apiClient := uc.cli.Client()
	return apiClient.ContainerUpdate(ctx, container, updateConfig)
}

// updateExample shows examples in update command, and is used in auto-generated cli docs.
func updateExample() string {
	return `$ pouch run -d -m 20m --name test-update registry.hub.docker.com/library/busybox:latest
8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812
$ cat /sys/fs/cgroup/memory/8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812/memory.limit_in_bytes
20971520
$ pouch update -m 30m test-update
$ cat /sys/fs/cgroup/memory/8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812/memory.limit_in_bytes
31457280
	`
}
