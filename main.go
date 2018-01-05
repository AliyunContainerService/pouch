package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/alibaba/pouch/daemon"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/exec"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/version"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfg          config.Config
	sigHandles   []func() error
	printVersion bool
)

func main() {
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
	flagSet.StringVar(&cfg.StreamServerAddress, "stream-addr", "", "The ip address streaming server of CRI is listening on. The default host interface is used if not specified.")
	flagSet.StringVar(&cfg.StreamServerPort, "stream-port", "10010", "The port streaming server of CRI is listening on.")
	flagSet.BoolVarP(&cfg.Debug, "debug", "D", false, "Switch daemon log level to DEBUG mode")
	flagSet.StringVarP(&cfg.ContainerdAddr, "containerd", "c", "/var/run/containerd.sock", "Specify listening address of containerd")
	flagSet.StringVar(&cfg.ContainerdPath, "containerd-path", "/usr/local/bin/containerd", "Specify the path of containerd binary")
	flagSet.StringVar(&cfg.ContainerdConfig, "containerd-config", "/etc/containerd/config.toml", "Specify the path of containerd configuration file")
	flagSet.StringVar(&cfg.TLS.Key, "tlskey", "", "Specify key file of TLS")
	flagSet.StringVar(&cfg.TLS.Cert, "tlscert", "", "Specify cert file of TLS")
	flagSet.StringVar(&cfg.TLS.CA, "tlscacert", "", "Specify CA file of TLS")
	flagSet.BoolVar(&cfg.TLS.VerifyRemote, "tlsverify", false, "Use TLS and verify remote")
	flagSet.BoolVarP(&printVersion, "version", "v", false, "Print daemon version")
	flagSet.StringVar(&cfg.DefaultRuntime, "default-runtime", "runc", "Default OCI Runtime")
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
	var processes exec.Processes = []*exec.Process{
		{
			Path: cfg.ContainerdPath,
			Args: []string{
				"-c", cfg.ContainerdConfig,
				"-a", cfg.ContainerdAddr,
				"--root", path.Join(cfg.HomeDir, "containerd/root"),
				"--state", path.Join(cfg.HomeDir, "containerd/state"),
				"-l", utils.If(cfg.Debug, "debug", "info").(string),
			},
		},
	}
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
