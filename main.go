package main

import (
	"fmt"
	"os"
	osexec "os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/alibaba/pouch/daemon"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/lxcfs"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/version"

	"github.com/docker/docker/pkg/reexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfg          config.Config
	sigHandles   []func() error
	printVersion bool
)

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

	if err := cmdServe.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

// setupFlags setups flags for command line.
func setupFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringVar(&cfg.HomeDir, "home-dir", "/var/lib/pouch", "Specify root dir of pouchd")
	flagSet.StringArrayVarP(&cfg.Listen, "listen", "l", []string{"unix:///var/run/pouchd.sock"}, "Specify listening addresses of Pouchd")
	flagSet.StringVar(&cfg.ListenCRI, "listen-cri", "/var/run/pouchcri.sock", "Specify listening address of CRI")
	flagSet.BoolVarP(&cfg.Debug, "debug", "D", false, "Switch daemon log level to DEBUG mode")
	flagSet.StringVarP(&cfg.ContainerdAddr, "containerd", "c", "/var/run/containerd.sock", "Specify listening address of containerd")
	flagSet.StringVar(&cfg.ContainerdPath, "containerd-path", "", "Specify the path of containerd binary")
	flagSet.StringVar(&cfg.TLS.Key, "tlskey", "", "Specify key file of TLS")
	flagSet.StringVar(&cfg.TLS.Cert, "tlscert", "", "Specify cert file of TLS")
	flagSet.StringVar(&cfg.TLS.CA, "tlscacert", "", "Specify CA file of TLS")
	flagSet.BoolVar(&cfg.TLS.VerifyRemote, "tlsverify", false, "Use TLS and verify remote")
	flagSet.BoolVarP(&printVersion, "version", "v", false, "Print daemon version")
	flagSet.StringVar(&cfg.DefaultRuntime, "default-runtime", "runc", "Default OCI Runtime")
	flagSet.BoolVar(&cfg.IsLxcfsEnabled, "enable-lxcfs", false, "Enable Lxcfs to make container to isolate /proc")
	flagSet.StringVar(&cfg.LxcfsBinPath, "lxcfs", "/usr/local/bin/lxcfs", "Specify the path of lxcfs binary")
	flagSet.StringVar(&cfg.LxcfsHome, "lxcfs-home", "/var/lib/lxc/lxcfs", "Specify the mount dir of lxcfs")
	flagSet.StringVar(&cfg.DefaultRegistry, "default-registry", "registry.hub.docker.com/library/", "Default Image Registry")
	flagSet.StringVar(&cfg.ImageProxy, "image-proxy", "http://127.0.0.1:5678", "Http proxy to pull image")
}

// runDaemon prepares configs, setups essential details and runs pouchd daemon.
func runDaemon() error {
	//user specifies --version or -v, print version and return.
	if printVersion {
		fmt.Println(version.Version)
		return nil
	}

	// initialize log.
	initLog()

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

	// define and start all required processes.
	if _, err := os.Stat(cfg.ContainerdAddr); err == nil {
		os.RemoveAll(cfg.ContainerdAddr)
	}

	containerdBinaryFile := "containerd"
	if cfg.ContainerdPath != "" {
		containerdBinaryFile = cfg.ContainerdPath
	}

	containerdPath, err := osexec.LookPath(containerdBinaryFile)
	if err != nil {
		return fmt.Errorf("failed to find containerd binary %s: %s", containerdBinaryFile, err)
	}

	var processes exec.Processes = []*exec.Process{
		{
			Path: containerdPath,
			Args: []string{
				"-a", cfg.ContainerdAddr,
				"--root", path.Join(cfg.HomeDir, "containerd/root"),
				"--state", path.Join(cfg.HomeDir, "containerd/state"),
				"-l", utils.If(cfg.Debug, "debug", "info").(string),
			},
		},
	}

	if err := checkLxcfsCfg(); err != nil {
		return err
	}
	processes = setLxcfsProcess(processes)
	defer processes.StopAll()

	if err := processes.RunAll(); err != nil {
		return err
	}
	sigHandles = append(sigHandles, processes.StopAll)

	// initialize signal and handle method.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		sig := <-signals
		logrus.Warnf("received signal: %s", sig)

		for _, handle := range sigHandles {
			if err := handle(); err != nil {
				logrus.Errorf("failed to handle signal: %v", err)
			}
		}
		os.Exit(1)
	}()

	// new daemon instance, this is core.
	d := daemon.NewDaemon(cfg)
	if d == nil {
		return fmt.Errorf("failed to new daemon")
	}

	sigHandles = append(sigHandles, d.Shutdown)

	return d.Run()
}

// initLog initializes log Level and log format of daemon.
func initLog() {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("start daemon at debug level")
	}

	formatter := &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
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
	cfg.LxcfsHome = strings.TrimSuffix(cfg.LxcfsHome, "/")

	lxcfs.IsLxcfsEnabled = cfg.IsLxcfsEnabled
	lxcfs.LxcfsHomeDir = cfg.LxcfsHome
	lxcfs.LxcfsParentDir = path.Dir(cfg.LxcfsHome)

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

	if _, err := os.Stat(cfg.LxcfsBinPath); err != nil {
		return fmt.Errorf("invalid lxcfs bin path: %s", cfg.LxcfsBinPath)
	}

	if _, err := os.Stat(cfg.LxcfsHome); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.LxcfsHome, 0755); err != nil {
				return fmt.Errorf("failed to LxcfsHome %s: %v", cfg.LxcfsHome, err)
			}
		}
	}
	return nil
}
