package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"

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

// Run is the entry of VolumeCommand command.
func (v *VolumeCommand) Run(args []string) {}

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
		Use:   "create [args] NAME",
		Short: "Create a pouch volume",
	}

	flagSet := v.cmd.Flags()

	flagSet.StringVarP(
		&v.name,
		"name",
		"n",
		"",
		"create volume name")
	flagSet.StringVarP(
		&v.driver,
		"driver",
		"d",
		"local",
		"create volume with driver")
	flagSet.StringSliceVarP(
		&v.options,
		"option",
		"o",
		nil,
		"create volume with options")
	flagSet.StringSliceVarP(
		&v.labels,
		"label",
		"l",
		nil,
		"create volume with labels")

	flagSet.StringSliceVarP(
		&v.selectors,
		"selector",
		"s",
		nil,
		"create volume with selectors")
}

// Run is the entry of VolumeCreateCommand command.
func (v *VolumeCreateCommand) Run(args []string) {
	logrus.Debugf("create a volume: %s, options: %v, labels: %v, selectors: %v",
		v.name, v.options, v.labels, v.selectors)
	v.volumeCreate()
}

func (v *VolumeCreateCommand) volumeCreate() error {
	volumeReq := &types.VolumeCreateRequest{
		Driver:     v.driver,
		Name:       v.name,
		DriverOpts: map[string]string{},
		Labels:     map[string]string{},
	}

	// analyze labels.
	for _, l := range v.labels {
		label := strings.Split(l, "=")
		if len(label) != 2 {
			err := fmt.Errorf("unknown label: %s", l)
			v.cli.Print(err)
			return err
		}
		volumeReq.Labels[label[0]] = label[1]
	}

	// analyze options.
	for _, o := range v.options {
		option := strings.Split(o, "=")
		if len(option) != 2 {
			err := fmt.Errorf("unknown option: %s", o)
			v.cli.Print(err)
			return err
		}
		volumeReq.DriverOpts[option[0]] = option[1]
	}

	// analyze selectors.
	for _, s := range v.selectors {
		selector := strings.Split(s, "=")
		if len(selector) != 2 {
			err := fmt.Errorf("unknown selector: %s", s)
			v.cli.Print(err)
			return err
		}
		volumeReq.DriverOpts["selector."+selector[0]] = selector[1]
	}

	client, err := client.New("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new client: %v\n", err)
		return err
	}

	volume, err := client.VolumeCreate(volumeReq)
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	v.cli.Print(volume)
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
		Use:     "remove [volume]",
		Aliases: []string{"rm"},
		Short:   "Remove pouch volumes",
		Args:    cobra.MinimumNArgs(1),
	}
}

// Run is the entry of VolumeRemoveCommand command.
func (v *VolumeRemoveCommand) Run(args []string) {
	name := args[0]

	logrus.Debugf("remove a volume: %s", name)

	client, err := client.New("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new client: %v\n", err)
		return
	}

	if err = client.VolumeRemove(name); err != nil {
		logrus.Errorln(err)
	}
}
