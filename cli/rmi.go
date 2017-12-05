package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RemoveImagesCommand use to implemt 'rmi' command, it remove one or more images by reference
type RmiCommand struct {
	baseCommand
	force bool
}

//Init initialize rmi command
func (cc *RmiCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "rmi image ",
		Short: "Remove one or more images by reference",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runRmi(args)
		},
	}
	cc.addFlags()
}

// addFlags adds flags for specific command
func (r *RmiCommand) addFlags() {
	r.cmd.Flags().BoolVarP(&r.force, "force", "f", false, "if image is inuse, remove image and all associated resources")
}

func (r *RmiCommand) runRmi(args []string) error {
	apiClient := r.cli.Client()

	for _, name := range args {
		if err := apiClient.ImageRemove(name, r.force); err != nil {
			return fmt.Errorf("failed to remove image: %v", err)
		}
		fmt.Printf("%s\n", name)
	}

	return nil
}
