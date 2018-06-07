package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/daemon"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/lxcfs"
	"github.com/alibaba/pouch/pkg/debug"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/storage/quota"
	"github.com/alibaba/pouch/version"

	"github.com/docker/docker/pkg/reexec"
	"github.com/google/gops/agent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sigHandles   []func() error
	printVersion bool
)

var cfg = &config.Config{}

func main() {
	if reexec.Init() {
		return
	}

	var cmdServe = &cobra.Command{
		Use:          "pouchd",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon()
		},
	}

	setupFlags(cmdServe)
	parseFlags(cmdServe, os.Args[1:])
	if err := loadDaemonFile(cfg, cmdServe.Flags()); err != nil {
		logrus.Errorf("failed to load daemon file: %s", err)
		os.Exit(1)
	}

	if err := cmdServe.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

// setupFlags setups flags for command line.
func setupFlags(cmd *cobra.Command) {
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	// flagSet := cmd.PersistentFlags()

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	flagSet := cmd.Flags()

	flagSet.StringVar(&cfg.HomeDir, "home-dir", "/var/lib/pouch", "Specify root dir of pouchd")
	flagSet.StringArrayVarP(&cfg.Listen, "listen", "l", []string{"unix:///var/run/pouchd.sock"}, "Specify listening addresses of Pouchd")
	flagSet.BoolVar(&cfg.IsCriEnabled, "enable-cri", false, "Specify whether enable the cri part of pouchd which is used to support Kubernetes")
	flagSet.StringVar(&cfg.CriConfig.CriVersion, "cri-version", "v1alpha2", "Specify the version of cri which is used to support Kubernetes")
	flagSet.StringVar(&cfg.CriConfig.Listen, "listen-cri", "/var/run/pouchcri.sock", "Specify listening address of CRI")
	flagSet.StringVar(&cfg.CriConfig.NetworkPluginBinDir, "cni-bin-dir", "/opt/cni/bin", "The directory for putting cni plugin binaries.")
	flagSet.StringVar(&cfg.CriConfig.NetworkPluginConfDir, "cni-conf-dir", "/etc/cni/net.d", "The directory for putting cni plugin configuration files.")
	flagSet.StringVar(&cfg.CriConfig.SandboxImage, "sandbox-image", "registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0", "The image used by sandbox container.")
	flagSet.BoolVarP(&cfg.Debug, "debug", "D", false, "Switch daemon log level to DEBUG mode")
	flagSet.StringVarP(&cfg.ContainerdAddr, "containerd", "c", "/var/run/containerd.sock", "Specify listening address of containerd")
	flagSet.StringVar(&cfg.ContainerdPath, "containerd-path", "", "Specify the path of containerd binary")
	flagSet.StringVar(&cfg.TLS.Key, "tlskey", "", "Specify key file of TLS")
	flagSet.StringVar(&cfg.TLS.Cert, "tlscert", "", "Specify cert file of TLS")
	flagSet.StringVar(&cfg.TLS.CA, "tlscacert", "", "Specify CA file of TLS")
	flagSet.BoolVar(&cfg.TLS.VerifyRemote, "tlsverify", false, "Use TLS and verify remote")
	flagSet.StringVar(&cfg.TLS.ManagerWhiteList, "manager-whitelist", "", "Set tls name whitelist, multiple values are separated by commas")
	flagSet.BoolVarP(&printVersion, "version", "v", false, "Print daemon version")
	flagSet.StringVar(&cfg.DefaultRuntime, "default-runtime", "runc", "Default OCI Runtime")
	flagSet.BoolVar(&cfg.IsLxcfsEnabled, "enable-lxcfs", false, "Enable Lxcfs to make container to isolate /proc")
	flagSet.StringVar(&cfg.LxcfsBinPath, "lxcfs", "/usr/local/bin/lxcfs", "Specify the path of lxcfs binary")
	flagSet.StringVar(&cfg.LxcfsHome, "lxcfs-home", "/var/lib/lxcfs", "Specify the mount dir of lxcfs")
	flagSet.StringVar(&cfg.DefaultRegistry, "default-registry", "registry.hub.docker.com", "Default Image Registry")
	flagSet.StringVar(&cfg.DefaultRegistryNS, "default-registry-namespace", "library", "Default Image Registry namespace")
	flagSet.StringVar(&cfg.ImageProxy, "image-proxy", "", "Set proxy to pull image. If not set, using HTTPS_PROXY, https_proxy, HTTP_PROXY, http_proxy by order")
	flagSet.StringVar(&cfg.QuotaDriver, "quota-driver", "", "Set quota driver(grpquota/prjquota), if not set, it will set by kernel version")
	flagSet.StringVar(&cfg.ConfigFile, "config-file", "/etc/pouch/config.json", "Configuration file of pouchd")

	// volume config
	flagSet.StringVar(&cfg.VolumeConfig.DriverAlias, "volume-driver-alias", "", "Set volume driver alias, <name=alias>[;name1=alias1]")

	// cgroup-path flag is to set parent cgroup for all containers, default is "default" staying with containerd's configuration.
	flagSet.StringVar(&cfg.CgroupParent, "cgroup-parent", "default", "Set parent cgroup for all containers")
	flagSet.StringVar(&cfg.PluginPath, "plugin", "", "Set the path where plugin shared library file put")
	flagSet.StringSliceVar(&cfg.Labels, "label", []string{}, "Set metadata for Pouch daemon")
	flagSet.BoolVar(&cfg.EnableProfiler, "enable-profiler", false, "Set if pouchd setup profiler")
	flagSet.StringVar(&cfg.Pidfile, "pidfile", "/var/run/pouch.pid", "Save daemon pid")
	flagSet.IntVar(&cfg.OOMScoreAdjust, "oom-score-adj", -500, "Set the oom_score_adj for the daemon")
}

// parse flags
func parseFlags(cmd *cobra.Command, flags []string) {
	err := cmd.Flags().Parse(flags)
	if err == nil || err == pflag.ErrHelp {
		return
	}

	cmd.SetOutput(os.Stderr)
	cmd.Usage()
}

// runDaemon prepares configs, setups essential details and runs pouchd daemon.
func runDaemon() error {
	//user specifies --version or -v, print version and return.
	if printVersion {
		fmt.Printf("pouchd version: %s, build: %s, build at: %s\n", version.Version, version.GitCommit, version.BuildTime)
		return nil
	}

	// initialize log.
	initLog()

	if err := cfg.Validate(); err != nil {
		logrus.Fatal(err)
	}

	// import debugger tools for pouch when in debug mode.
	if cfg.Debug || cfg.EnableProfiler {
		if err := agent.Listen(agent.Options{}); err != nil {
			logrus.Fatal(err)
		}
		debug.SetupDumpStackTrap()
	}

	// initialize home dir.
	dir := cfg.HomeDir

	if dir == "" || !path.IsAbs(dir) {
		return fmt.Errorf("invalid pouchd's home dir: %s", dir)
	}
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0666); err != nil {
			return fmt.Errorf("failed to mkdir: %v", err)
		}
	}

	// saves daemon pid to pidfile.
	if cfg.Pidfile != "" {
		if err := utils.NewPidfile(cfg.Pidfile); err != nil {
			logrus.Errorf("failed to create pidfile: %s", err)
			return err
		}
		defer func() {
			if err := os.Remove(cfg.Pidfile); err != nil {
				logrus.Errorf("failed to delete pidfile: %s", err)
			}
		}()
	}

	// set pouchd oom-score
	if err := utils.SetOOMScore(os.Getpid(), cfg.OOMScoreAdjust); err != nil {
		logrus.Errorf("failed to set oom-score for pouchd: %v", err)
	}

	// define and start all required processes.

	if cfg.QuotaDriver != "" {
		quota.SetQuotaDriver(cfg.QuotaDriver)
	}

	if err := checkLxcfsCfg(); err != nil {
		return err
	}

	// initialize signal and handle method.
	var (
		errCh    = make(chan error, 1)
		signalCh = make(chan os.Signal, 1)
	)

	// new daemon instance, this is core.
	d := daemon.NewDaemon(cfg)
	if d == nil {
		return fmt.Errorf("failed to new daemon")
	}

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	sigHandles = append(sigHandles, d.ShutdownPlugin, d.Shutdown)

	go func() {
		// FIXME: I think the Run() should always return error.
		errCh <- d.Run()
	}()

	select {
	case sig := <-signalCh:
		logrus.Warnf("received signal: %s", sig)

		for _, handle := range sigHandles {
			if err := handle(); err != nil {
				logrus.Errorf("failed to handle signal: %v", err)
			}
		}

		os.Exit(1)
	case err := <-errCh:
		// FIXME: should we do the cleanup like signal handle?
		return err
	}
	return nil
}

// initLog initializes log Level and log format of daemon.
func initLog() {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("start daemon at debug level")
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logrus.SetFormatter(formatter)
}

// define lxcfs processe.
func setLxcfsProcess(processes exec.Processes) exec.Processes {
	if !cfg.IsLxcfsEnabled {
		return processes
	}

	p := &exec.Process{
		Path: cfg.LxcfsBinPath,
		Args: []string{
			cfg.LxcfsHome,
		},
	}
	processes = append(processes, p)

	return processes
}

// check lxcfs config
func checkLxcfsCfg() error {
	if !cfg.IsLxcfsEnabled {
		return nil
	}

	if !path.IsAbs(cfg.LxcfsHome) {
		return fmt.Errorf("invalid lxcfs home dir: %s", cfg.LxcfsHome)
	}

	cfg.LxcfsHome = strings.TrimSuffix(cfg.LxcfsHome, "/")
	lxcfs.IsLxcfsEnabled = cfg.IsLxcfsEnabled
	lxcfs.LxcfsHomeDir = cfg.LxcfsHome
	lxcfs.LxcfsParentDir = path.Dir(cfg.LxcfsHome)

	return lxcfs.CheckLxcfsMount()
}

// load daemon config file
func loadDaemonFile(cfg *config.Config, flagSet *pflag.FlagSet) error {
	if cfg.ConfigFile == "" {
		return nil
	}

	return cfg.MergeConfigurations(cfg, flagSet)
}
