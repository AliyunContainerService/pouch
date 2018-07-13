package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

// checkpointDescription is used to describe checkpoint command in detail and auto generate command doc.
var checkpointDescription = "\nManage checkpoint commands, create checkpoint."

// CheckpointCommand use to implement 'checkpoint' command, it checkpoint a container.
type CheckpointCommand struct {
	baseCommand
}

// Init initialize checkpoint command.
func (cp *CheckpointCommand) Init(c *Cli) {
	cp.cli = c
	cp.cmd = &cobra.Command{
		Use:   "checkpoint COMMAND",
		Short: "Manage checkpoint commands",
		Long:  checkpointDescription,
		Args:  cobra.MinimumNArgs(1),
	}

	// add subcommands
	c.AddCommand(cp, &CheckpointCreateCommand{})
	c.AddCommand(cp, &CheckpointListCommand{})
	c.AddCommand(cp, &CheckpointDelCommand{})
}

// checkpoint subcommands

// checkpointCreateDescription is used to describe checkpoint create command in detail and auto generate command doc.
var checkpointCreateDescription = "Create a checkpoint from a running container instance keep the state for restore later."

// CheckpointCreateCommand use to implement 'checkpoint create' command, it create a container checkpoint.
type CheckpointCreateCommand struct {
	CheckpointCommand

	leaveRunning bool
	cpDir        string
}

// Init initialize checkpoint create command.
func (cc *CheckpointCreateCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "create [OPTIONS] CONTAINER CHECKPOINT",
		Short: "create a checkpoint from a running container instance",
		Long:  checkpointCreateDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCheckpointCreate(args)
		},
		Example: checkpointCreateExample(),
	}
	cc.addFlags()
}

// runCheckpoint is the entry of checkpoint create command.
func (cc *CheckpointCreateCommand) runCheckpointCreate(args []string) error {
	ctx := context.Background()
	apiClient := cc.cli.Client()

	if err := apiClient.ContainerCheckpointCreate(ctx, args[0], types.CheckpointCreateOptions{
		CheckpointID:  args[1],
		CheckpointDir: cc.cpDir,
		Exit:          !cc.leaveRunning,
	}); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, args[1])
	return nil
}

// addFlags adds flags for specific command.
func (cc *CheckpointCreateCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.BoolVar(&cc.leaveRunning, "leave-running", false, "keep container running after creating checkpoint")
	flagSet.StringVar(&cc.cpDir, "checkpoint-dir", "", "directory to store checkpoints images")
}

// checkpointCreateExample shows examples in checkpoint create command, and is used in auto-generated cli docs.
func checkpointCreateExample() string {
	return `$ pouch checkpoint create container-name cp0
cp0`
}

// checkpointListDescription is used to describe checkpoint list command in detail and auto generate command doc.
var checkpointListDescription = "List a container checkpoint."

// CheckpointListCommand use to implement 'checkpoint list' command, it list a container checkpoint.
type CheckpointListCommand struct {
	CheckpointCommand
	cpDir string
}

// Init initialize checkpoint list command.
func (cc *CheckpointListCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "ls [OPTIONS] CONTAINER",
		Short: "list checkpoints of a container",
		Long:  checkpointListDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCheckpointList(args)
		},
		Example: checkpointListExample(),
	}
	cc.addFlags()
}

// runCheckpoint is the entry of checkpoint list command.
func (cc *CheckpointListCommand) runCheckpointList(args []string) error {
	ctx := context.Background()
	apiClient := cc.cli.Client()

	list, err := apiClient.ContainerCheckpointList(ctx, args[0], types.CheckpointListOptions{
		CheckpointDir: cc.cpDir,
	})
	if err != nil {
		return err
	}

	for _, cp := range list {
		fmt.Fprintln(os.Stdout, cp)
	}
	return nil
}

// addFlags adds flags for specific command.
func (cc *CheckpointListCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.StringVar(&cc.cpDir, "checkpoint-dir", "", "directory to store checkpoints images")
}

// checkpointListExample shows examples in checkpoint list command, and is used in auto-generated cli docs.
func checkpointListExample() string {
	return `$ pouch checkpoint list container-name
cp0`
}

// checkpointDelDescription is used to describe checkpoint delete command in detail and auto generate command doc.
var checkpointDelDescription = "Delete a container checkpoint."

// CheckpointDelCommand use to implement 'checkpoint delete' command, it delete a container checkpoint.
type CheckpointDelCommand struct {
	CheckpointCommand
	cpDir string
}

// Init initialize checkpoint delete command.
func (cc *CheckpointDelCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "rm [OPTIONS] CONTAINER CHECKPOINT",
		Short: "delete a container checkpoint",
		Long:  checkpointDelDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCheckpointDelete(args)
		},
		Example: checkpointDeleteExample(),
	}
	cc.addFlags()
}

// runCheckpoint is the entry of checkpoint delete command.
func (cc *CheckpointDelCommand) runCheckpointDelete(args []string) error {
	ctx := context.Background()
	apiClient := cc.cli.Client()

	if err := apiClient.ContainerCheckpointDelete(ctx, args[0], types.CheckpointDeleteOptions{
		CheckpointID:  args[1],
		CheckpointDir: cc.cpDir,
	}); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, args[1])
	return nil
}

// addFlags adds flags for specific command.
func (cc *CheckpointDelCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.StringVar(&cc.cpDir, "checkpoint-dir", "", "directory to store checkpoints images")
}

// checkpointDeleteExample shows examples in checkpoint delete command, and is used in auto-generated cli docs.
func checkpointDeleteExample() string {
	return `$ pouch checkpoint delete container-name
cp0`
}
