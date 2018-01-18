package main

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// volumeDescription is used to describe volume command in detail and auto generate command doc.
var volumeDescription = "Manager the volumes in pouchd. " +
	"It contains the functions of create/remove/list/info volume, 'driver' is used to list drivers that pouch support. " +
	"The default volume driver is local, it will make a directory to bind into container."

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
		Long:  volumeDescription,
		Args:  cobra.MinimumNArgs(1),
	}

	c.AddCommand(v, &VolumeCreateCommand{})
	c.AddCommand(v, &VolumeRemoveCommand{})
	c.AddCommand(v, &VolumeInfoCommand{})
}

// RunE is the entry of VolumeCommand command.
func (v *VolumeCommand) RunE(args []string) error {
	return nil
}

// volumeCreateDescription is used to describe volume create command in detail and auto generate command doc.
var volumeCreateDescription = "Create a volume in pouchd. " +
	"It must specify volume's name, size and driver. You can use 'volume driver' to get drivers that pouch support."

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
		Use:   "create [OPTIONS]",
		Short: "Create a volume",
		Long:  volumeCreateDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVolumeCreate(args)
		},
		Example: volumeCreateExample(),
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
	volumeReq := &types.VolumeCreateConfig{
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

func parseVolume(volumeCreateConfig *types.VolumeCreateConfig, v *VolumeCreateCommand) error {
	// analyze labels.
	for _, label := range v.labels {
		l := strings.Split(label, "=")
		if len(label) != 2 {
			return fmt.Errorf("unknown label %s: label format must be key=value", label)
		}
		volumeCreateConfig.Labels[l[0]] = l[1]
	}

	// analyze options.
	for _, option := range v.options {
		opt := strings.Split(option, "=")
		if len(opt) != 2 {
			return fmt.Errorf("unknown option %s: option format must be key=value", option)
		}
		volumeCreateConfig.DriverOpts[opt[0]] = opt[1]
	}

	// analyze selectors.
	for _, selector := range v.selectors {
		s := strings.Split(selector, "=")
		if len(s) != 2 {
			return fmt.Errorf("unknown selector %s: selector format must be key=value", selector)
		}
		volumeCreateConfig.DriverOpts["selector."+s[0]] = s[1]
	}
	return nil
}

// volumeCreateExample shows examples in volume create command, and is used in auto-generated cli docs.
func volumeCreateExample() string {
	return `$ pouch volume create -d local -n pouch-volume -o size=100g
Mountpoint:
Name:         pouch-volume
Scope:
CreatedAt:
Driver:       local`
}

// volumeRmDescription is used to describe volume rm command in detail and auto generate command doc.
var volumeRmDescription = "Remove a volume in pouchd. " +
	"It need specify volume's name, when the volume is exist and is unuse, it will be remove."

// VolumeRemoveCommand is used to implement 'volume rm' command.
type VolumeRemoveCommand struct {
	baseCommand
}

// Init initializes VolumeRemoveCommand command.
func (v *VolumeRemoveCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:     "remove VOLUME",
		Aliases: []string{"rm"},
		Short:   "Remove volume",
		Long:    volumeRmDescription,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVolumeRm(args)
		},
		Example: volumeRmExample(),
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VolumeRemoveCommand) addFlags() {}

// runVolumeRm is the entry of VolumeRemoveCommand command.
func (v *VolumeRemoveCommand) runVolumeRm(args []string) error {
	name := args[0]

	logrus.Debugf("remove a volume: %s", name)

	apiClient := v.cli.Client()

	err := apiClient.VolumeRemove(name)
	if err == nil {
		fmt.Printf("Removed: %s\n", name)
	}

	return err
}

// volumeRmExample shows examples in volume rm command, and is used in auto-generated cli docs.
func volumeRmExample() string {
	return `$ pouch volume rm pouch-volume
Removed: pouch-volume`
}

// volumeInfoDescription is used to describe volume information and auto generate command doc.
var volumeInfoDescription = "Get a volume's information. " +
	"It need specify volume's name, and then return the volume's information that include driver, mountpoint and so on."

// VolumeInfoCommand is used to implement 'volume info' command.
type VolumeInfoCommand struct {
	baseCommand
}

// Init initializes VolumeInfoCommand command.
func (v *VolumeInfoCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:     "info VOLUME",
		Aliases: []string{"info"},
		Short:   "info volume",
		Long:    volumeInfoDescription,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVolumeInfo(args)
		},
		Example: volumeInfoExample(),
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VolumeInfoCommand) addFlags() {}

// runVolumeInfo is the entry of VolumeInfoCommand command.
func (v *VolumeInfoCommand) runVolumeInfo(args []string) error {
	name := args[0]

	logrus.Debugf("get a the information of volume: %s", name)

	apiClient := v.cli.Client()

	volume, err := apiClient.VolumeInfo(name)
	if err != nil {
		return err
	}

	v.cli.Print(volume)
	return err
}

// volumeInfoExample shows examples in volume info command, and is used in auto-generated cli docs.
func volumeInfoExample() string {
	return `$ pouch volume info pouch-volume
Name:         pouch-volume
Scope:
CreatedAt:    2018-1-15 10:50:36
Driver:       local
Mountpoint:   /mnt/local/pouch-volume`
}
