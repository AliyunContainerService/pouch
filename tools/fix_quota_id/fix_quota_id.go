package main

import (
	"os"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/storage/quota"

	"github.com/spf13/cobra"
)

var (
	dir       string
	size      string
	quotaID   uint32
	recursive bool
)

func run(cmd *cobra.Command) error {
	_, err := quota.StartQuotaDriver(dir)
	if err != nil {
		log.With(nil).Errorf("failed to start quota driver for %s, err: %v", dir, err)
		return err
	}

	err = quota.SetDiskQuota(dir, size, quotaID)
	if err != nil {
		log.With(nil).Errorf("failed to set subtree for %s, quota id: %d, err: %v", dir, quotaID, err)
		return err
	}

	if !recursive {
		return nil
	}

	if err := quota.SetQuotaForDir(dir, quotaID); err != nil {
		log.With(nil).Errorf("failed to set quota id for %s recursively, quota id: %d, err: %v", dir, quotaID, err)
		return err
	}

	return nil
}

func setupFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringVarP(&dir, "dir", "d", "", "The directory is set quota id.")
	flagSet.StringVarP(&size, "size", "s", "", "The limit of directory")
	flagSet.Uint32VarP(&quotaID, "quota-id", "i", 0, "The quota id to set.")
	flagSet.BoolVarP(&recursive, "recursive", "r", true, "Set the directory recursively.")
}

func main() {
	var cmdServe = &cobra.Command{
		Use:          "fix_quota_id <-d /path> <-s size> <-i id>",
		Short:        "Set the quota id of path",
		Long:         "Set the quota id of path, also can set its sub-directory by recursively",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}

	setupFlags(cmdServe)
	if err := cmdServe.Execute(); err != nil {
		log.With(nil).Error(err)
		os.Exit(1)
	}
}
