package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/alibaba/pouch/daemon"
	"github.com/alibaba/pouch/daemon/config"
)

func main() {
	var cfg config.Config
	var cmdServe = &cobra.Command{
		Use:  "",
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.Debug {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.Debug("start daemon at debug level")
			}
			d := daemon.NewDaemon(cfg)
			if d == nil {
				os.Exit(1)
			}
			err := d.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		},
	}

	flagSet := cmdServe.Flags()
	flagSet.StringArrayVarP(&cfg.Listen, "listen", "l", []string{"unix:///var/run/pouchd.sock"}, "which address to listen on")
	flagSet.BoolVarP(&cfg.Debug, "debug", "D", false, "switch debug level")
	flagSet.StringVarP(&cfg.ContainerdAddr, "containerd", "c", "/var/run/containerd.sock", "where does containerd listened on")

	cmdServe.Execute()
}
