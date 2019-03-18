package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// daemonUpdateDescription is used to describe updatedaemon command in detail and auto generate command doc.
var daemonUpdateDescription = "Update daemon's configurations, if daemon is stoped, it will just update config file. " +
	"Online update just including: image proxy, label, offline update including: manager white list, debug level, " +
	"execute root directory, bridge name, bridge IP, fixed CIDR, defaut gateway, iptables, ipforwark, userland proxy. " +
	"If pouchd is alive, you can only use --offline=true to update config file"

// DaemonUpdateCommand use to implement 'updatedaemon' command, it modifies the configurations of a container.
type DaemonUpdateCommand struct {
	baseCommand

	configFile string
	offline    bool

	debug            bool
	imageProxy       string
	label            []string
	managerWhiteList string
	execRoot         string
	disableBridge    bool
	bridgeName       string
	bridgeIP         string
	fixedCIDRv4      string
	defaultGatewayv4 string
	iptables         bool
	ipforward        bool
	userlandProxy    bool

	homeDir     string
	snapshotter string
}

// Init initialize updatedaemon command.
func (udc *DaemonUpdateCommand) Init(c *Cli) {
	udc.cli = c
	udc.cmd = &cobra.Command{
		Use:   "updatedaemon [OPTIONS]",
		Short: "Update the configurations of pouchd",
		Long:  daemonUpdateDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return udc.daemonUpdateRun(args)
		},
		Example: daemonUpdateExample(),
	}
	udc.addFlags()
}

// addFlags adds flags for specific command.
func (udc *DaemonUpdateCommand) addFlags() {
	flagSet := udc.cmd.Flags()
	flagSet.SetInterspersed(false)
	flagSet.StringVar(&udc.configFile, "config-file", "/etc/pouch/config.json", "specified config file for updating daemon")
	flagSet.BoolVar(&udc.offline, "offline", false, "just update daemon config file")

	flagSet.BoolVar(&udc.debug, "debug", false, "update daemon debug mode")
	flagSet.StringVar(&udc.imageProxy, "image-proxy", "", "update daemon image proxy")
	flagSet.StringVar(&udc.managerWhiteList, "manager-white-list", "", "update daemon manager white list")
	flagSet.StringSliceVar(&udc.label, "label", nil, "update daemon labels")
	flagSet.StringVar(&udc.execRoot, "exec-root-dir", "", "update exec root directory for network")
	flagSet.BoolVar(&udc.disableBridge, "disable-bridge", false, "disable bridge network")
	flagSet.StringVar(&udc.bridgeName, "bridge-name", "", "update daemon bridge device")
	flagSet.StringVar(&udc.bridgeIP, "bip", "", "update daemon bridge IP")
	flagSet.StringVar(&udc.fixedCIDRv4, "fixed-cidr", "", "update daemon bridge fixed CIDR")
	flagSet.StringVar(&udc.defaultGatewayv4, "default-gateway", "", "update daemon bridge default gateway")
	flagSet.BoolVar(&udc.iptables, "iptables", true, "update daemon with iptables")
	flagSet.BoolVar(&udc.ipforward, "ipforward", true, "udpate daemon with ipforward")
	flagSet.BoolVar(&udc.userlandProxy, "userland-proxy", false, "update daemon with userland proxy")
	flagSet.StringVar(&udc.homeDir, "home-dir", "", "update daemon home dir")
	flagSet.StringVar(&udc.snapshotter, "snapshotter", "", "update daemon snapshotter")
}

// daemonUpdateRun is the entry of updatedaemon command.
func (udc *DaemonUpdateCommand) daemonUpdateRun(args []string) error {
	ctx := context.Background()

	apiClient := udc.cli.Client()

	msg, err := apiClient.SystemPing(ctx)
	if !udc.offline && err == nil && msg == "OK" {
		// TODO: daemon support more configures for update online, such as debug level.
		daemonConfig := &types.DaemonUpdateConfig{
			ImageProxy: udc.imageProxy,
			Labels:     udc.label,
		}

		err = apiClient.DaemonUpdate(ctx, daemonConfig)
		if err != nil {
			return errors.Wrap(err, "failed to update alive daemon config")
		}
	} else {
		// offline update config file.
		err = udc.updateDaemonConfigFile()
		if err != nil {
			return errors.Wrap(err, "failed to update daemon config file.")
		}
	}

	return nil
}

// updateDaemonConfigFile is just used to update config file.
func (udc *DaemonUpdateCommand) updateDaemonConfigFile() error {
	// read config from file.
	contents, err := ioutil.ReadFile(udc.configFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read config file(%s)", udc.configFile)
	}

	daemonConfig := &config.Config{}
	// do not return error if config file is empty
	if err := json.NewDecoder(bytes.NewReader(contents)).Decode(daemonConfig); err != nil && err != io.EOF {
		return errors.Wrapf(err, "failed to decode json: %s", udc.configFile)
	}

	flagSet := udc.cmd.Flags()

	if flagSet.Changed("image-proxy") {
		daemonConfig.ImageProxy = udc.imageProxy
	}

	if flagSet.Changed("manager-white-list") {
		daemonConfig.TLS.ManagerWhiteList = udc.managerWhiteList
	}

	// TODO: add parse labels

	if flagSet.Changed("exec-root-dir") {
		daemonConfig.NetworkConfig.ExecRoot = udc.execRoot
	}

	if flagSet.Changed("disable-bridge") {
		daemonConfig.NetworkConfig.BridgeConfig.DisableBridge = udc.disableBridge
	}

	if flagSet.Changed("bridge-name") {
		daemonConfig.NetworkConfig.BridgeConfig.Name = udc.bridgeName
	}

	if flagSet.Changed("bip") {
		daemonConfig.NetworkConfig.BridgeConfig.IPv4 = udc.bridgeIP
	}

	if flagSet.Changed("fixed-cidr") {
		daemonConfig.NetworkConfig.BridgeConfig.FixedCIDRv4 = udc.fixedCIDRv4
	}

	if flagSet.Changed("default-gateway") {
		daemonConfig.NetworkConfig.BridgeConfig.GatewayIPv4 = udc.defaultGatewayv4
	}

	if flagSet.Changed("iptables") {
		daemonConfig.NetworkConfig.BridgeConfig.IPTables = udc.iptables
	}

	if flagSet.Changed("ipforward") {
		daemonConfig.NetworkConfig.BridgeConfig.IPForward = udc.ipforward
	}

	if flagSet.Changed("userland-proxy") {
		daemonConfig.NetworkConfig.BridgeConfig.UserlandProxy = udc.userlandProxy
	}

	if flagSet.Changed("home-dir") {
		daemonConfig.HomeDir = udc.homeDir
	}

	if flagSet.Changed("snapshotter") {
		daemonConfig.Snapshotter = udc.snapshotter
	}

	// write config to file
	fd, err := os.OpenFile(udc.configFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open config file(%s)", udc.configFile)
	}
	defer fd.Close()

	fd.Seek(0, io.SeekStart)
	encoder := json.NewEncoder(fd)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(daemonConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to write config to file(%s)", udc.configFile)
	}

	return nil
}

// daemonUpdateExample shows examples in updatedaemon command, and is used in auto-generated cli docs.
func daemonUpdateExample() string {
	return `$ pouch updatedaemon --debug=true`
}
