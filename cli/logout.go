package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba/pouch/credential"

	"github.com/spf13/cobra"
)

// logoutDescription is used to describe logout command and auto generate command doc.
var logoutDescription = "\nlogout from a v1/v2 registry."

// LogoutCommand use to implement 'logout' command.
type LogoutCommand struct {
	baseCommand
}

// Init initialize logout command.
func (l *LogoutCommand) Init(c *Cli) {
	l.cli = c

	l.cmd = &cobra.Command{
		Use:   "logout [SERVER]",
		Short: "Logout from a registry",
		Long:  logoutDescription,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return l.runLogout(args)
		},
		Example: logoutExample(),
	}
}

// runLogout is the entry of logout command.
func (l *LogoutCommand) runLogout(args []string) error {
	var registry string
	if len(args) > 0 {
		registry = args[0]
	}

	if registry == "" {
		ctx := context.Background()
		info, err := l.cli.Client().SystemInfo(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fail to get default registry: %s\n", err)
			return err
		}
		registry = info.DefaultRegistry
	}

	if !credential.Exist(registry) {
		fmt.Fprintf(os.Stdout, "Has not logged in registry: %s\n", registry)
		return nil
	}

	if err := credential.Delete(registry); err != nil {
		fmt.Fprintf(os.Stderr, "Fail to remove login credential: %s\n", err)
		return err
	}

	fmt.Fprintf(os.Stdout, "Remove login credential for registry: %s\n", registry)
	return nil
}

// logoutExample shows examples in logout command, and is used in auto-generated cli docs.
func logoutExample() string {
	return `$ pouch logout $registry
Remove login credential for registry: $registry`
}
