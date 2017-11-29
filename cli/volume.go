package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// VolumeCommand is used to implement 'volume' command.
type VolumeCommand struct {
	baseCommand
}

// Init initializes VolumeCommand command.
func (v *VolumeCommand) Init(c *Cli) {
	v.cli = c

	v.cmd = &cobra.Command{
		Use:   "volume [command]",
		Short: "Manage pouch volumes",
		Args:  cobra.MinimumNArgs(1),
	}

	c.AddCommand(v, &VolumeCreateCommand{})
	c.AddCommand(v, &VolumeRemoveCommand{})
}

// RunE is the entry of VolumeCommand command.
func (v *VolumeCommand) RunE(args []string) error {
	return nil
}

// VolumeCreateCommand is used to implement 'volume create' command.
type VolumeCreateCommand struct {
	baseCommand

	name      string
	driver    string
	options   []string
	labels    []string
	selectors []string
}

// Init initializes VolumeCreateCommand command.
func (v *VolumeCreateCommand) Init(c *Cli) {
	v.cli = c

	v.cmd = &cobra.Command{
		Use:   "create [args]",
		Short: "Create a volume",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVolumeCreate(args)
		},
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VolumeCreateCommand) addFlags() {
	flagSet := v.cmd.Flags()
	flagSet.StringVarP(&v.name, "name", "n", "", "Specify name for volume")
	flagSet.StringVarP(&v.driver, "driver", "d", "local", "Specify volume driver name (default 'local')")
	flagSet.StringSliceVarP(&v.options, "option", "o", nil, "Set volume driver options")
	flagSet.StringSliceVarP(&v.labels, "label", "l", nil, "Set labels for volume")
	flagSet.StringSliceVarP(&v.selectors, "selector", "s", nil, "Set volume selectors")
}

// runVolumeCreate is the entry of VolumeCreateCommand command.
func (v *VolumeCreateCommand) runVolumeCreate(args []string) error {
	logrus.Debugf("create a volume: %s, options: %v, labels: %v, selectors: %v",
		v.name, v.options, v.labels, v.selectors)
	return v.volumeCreate()
}

func (v *VolumeCreateCommand) volumeCreate() error {
	volumeReq := &types.VolumeCreateRequest{
		Driver:     v.driver,
		Name:       v.name,
		DriverOpts: map[string]string{},
		Labels:     map[string]string{},
	}

	if err := parseVolume(volumeReq, v); err != nil {
		return err
	}

	apiClient := v.cli.Client()
	volume, err := apiClient.VolumeCreate(volumeReq)
	if err != nil {
		return err
	}

	v.cli.Print(volume)
	return nil
}

func parseVolume(volumeReq *types.VolumeCreateRequest, v *VolumeCreateCommand) error {
	// analyze labels.
	for _, label := range v.labels {
		l := strings.Split(label, "=")
		if len(label) != 2 {
			return fmt.Errorf("unknown label %s: label format must be key=value", label)
		}
		volumeReq.Labels[l[0]] = l[1]
	}

	// analyze options.
	for _, option := range v.options {
		opt := strings.Split(option, "=")
		if len(opt) != 2 {
			return fmt.Errorf("unknown option %s: option format must be key=value", option)
		}
		volumeReq.DriverOpts[opt[0]] = opt[1]
	}

	// analyze selectors.
	for _, selector := range v.selectors {
		s := strings.Split(selector, "=")
		if len(s) != 2 {
			return fmt.Errorf("unknown selector %s: selector format must be key=value", selector)
		}
		volumeReq.DriverOpts["selector."+s[0]] = s[1]
	}
	return nil
}

// VolumeRemoveCommand is used to implement 'volume rm' command.
type VolumeRemoveCommand struct {
	baseCommand
}

// Init initializes VolumeRemoveCommand command.
func (v *VolumeRemoveCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove volume",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVolumeRm(args)
		},
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VolumeRemoveCommand) addFlags() {
	// TODO: add flags here
}

// runVolumeRm is the entry of VolumeRemoveCommand command.
func (v *VolumeRemoveCommand) runVolumeRm(args []string) error {
	name := args[0]

	logrus.Debugf("remove a volume: %s", name)

	apiClient := v.cli.Client()

	return apiClient.VolumeRemove(name)
}
